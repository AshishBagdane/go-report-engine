package processor

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- Mock Processors for Testing ---

type mockProcessorHandler struct {
	BaseProcessor
	processFunc func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error)
	closeFunc   func() error
}

func (m *mockProcessorHandler) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, data)
	}
	return m.BaseProcessor.Process(ctx, data)
}

func (m *mockProcessorHandler) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type mockConfigurableProcessor struct {
	BaseProcessor
	configureFunc func(params map[string]string) error
	configured    bool
	params        map[string]string
}

func (m *mockConfigurableProcessor) Configure(params map[string]string) error {
	m.configured = true
	m.params = params
	if m.configureFunc != nil {
		return m.configureFunc(params)
	}
	return nil
}

func (m *mockConfigurableProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	return m.BaseProcessor.Process(ctx, data)
}

// --- Constructor Tests ---

func TestNewParallelProcessor(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)

	if parallel == nil {
		t.Fatal("NewParallelProcessor returned nil")
	}

	if parallel.processor != processor {
		t.Error("Wrapped processor not set correctly")
	}

	if parallel.workers != runtime.NumCPU() {
		t.Errorf("Default workers = %d, expected %d", parallel.workers, runtime.NumCPU())
	}

	if parallel.minChunkSize != 100 {
		t.Errorf("Default minChunkSize = %d, expected 100", parallel.minChunkSize)
	}

	if parallel.chunkSize != 0 {
		t.Errorf("Default chunkSize = %d, expected 0 (auto)", parallel.chunkSize)
	}
}

func TestNewParallelProcessor_NilProcessorPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewParallelProcessor with nil processor should panic")
		}
	}()

	NewParallelProcessor(nil)
}

func TestNewParallelProcessorWithConfig(t *testing.T) {
	processor := &BaseProcessor{}

	config := ParallelConfig{
		Workers:      8,
		ChunkSize:    500,
		MinChunkSize: 50,
	}

	parallel := NewParallelProcessorWithConfig(processor, config)

	if parallel == nil {
		t.Fatal("NewParallelProcessorWithConfig returned nil")
	}

	if parallel.workers != 8 {
		t.Errorf("Workers = %d, expected 8", parallel.workers)
	}

	if parallel.chunkSize != 500 {
		t.Errorf("ChunkSize = %d, expected 500", parallel.chunkSize)
	}

	if parallel.minChunkSize != 50 {
		t.Errorf("MinChunkSize = %d, expected 50", parallel.minChunkSize)
	}
}

func TestNewParallelProcessorWithConfig_DefaultValidation(t *testing.T) {
	processor := &BaseProcessor{}

	tests := []struct {
		name   string
		config ParallelConfig
		check  func(*ParallelProcessor) error
	}{
		{
			name: "zero workers uses CPU count",
			config: ParallelConfig{
				Workers:      0,
				MinChunkSize: 100,
			},
			check: func(p *ParallelProcessor) error {
				if p.workers != runtime.NumCPU() {
					return fmt.Errorf("workers = %d, expected %d", p.workers, runtime.NumCPU())
				}
				return nil
			},
		},
		{
			name: "negative workers uses CPU count",
			config: ParallelConfig{
				Workers:      -5,
				MinChunkSize: 100,
			},
			check: func(p *ParallelProcessor) error {
				if p.workers != runtime.NumCPU() {
					return fmt.Errorf("workers = %d, expected %d", p.workers, runtime.NumCPU())
				}
				return nil
			},
		},
		{
			name: "zero min chunk size uses default",
			config: ParallelConfig{
				Workers:      4,
				MinChunkSize: 0,
			},
			check: func(p *ParallelProcessor) error {
				if p.minChunkSize != 100 {
					return fmt.Errorf("minChunkSize = %d, expected 100", p.minChunkSize)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parallel := NewParallelProcessorWithConfig(processor, tt.config)
			if err := tt.check(parallel); err != nil {
				t.Error(err)
			}
		})
	}
}

// --- Configuration Tests ---

func TestParallelProcessorConfigure(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)

	params := map[string]string{
		"workers":        "16",
		"chunk_size":     "1000",
		"min_chunk_size": "50",
	}

	err := parallel.Configure(params)
	if err != nil {
		t.Fatalf("Configure() returned error: %v", err)
	}

	if parallel.Workers() != 16 {
		t.Errorf("Workers = %d, expected 16", parallel.Workers())
	}

	if parallel.ChunkSize() != 1000 {
		t.Errorf("ChunkSize = %d, expected 1000", parallel.ChunkSize())
	}

	if parallel.minChunkSize != 50 {
		t.Errorf("MinChunkSize = %d, expected 50", parallel.minChunkSize)
	}
}

func TestParallelProcessorConfigure_InvalidParams(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)

	tests := []struct {
		name   string
		params map[string]string
		errMsg string
	}{
		{
			name:   "invalid workers",
			params: map[string]string{"workers": "abc"},
			errMsg: "invalid workers parameter",
		},
		{
			name:   "zero workers",
			params: map[string]string{"workers": "0"},
			errMsg: "workers must be > 0",
		},
		{
			name:   "negative workers",
			params: map[string]string{"workers": "-5"},
			errMsg: "workers must be > 0",
		},
		{
			name:   "invalid chunk_size",
			params: map[string]string{"chunk_size": "xyz"},
			errMsg: "invalid chunk_size parameter",
		},
		{
			name:   "negative chunk_size",
			params: map[string]string{"chunk_size": "-100"},
			errMsg: "chunk_size must be >= 0",
		},
		{
			name:   "invalid min_chunk_size",
			params: map[string]string{"min_chunk_size": "def"},
			errMsg: "invalid min_chunk_size parameter",
		},
		{
			name:   "zero min_chunk_size",
			params: map[string]string{"min_chunk_size": "0"},
			errMsg: "min_chunk_size must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parallel.Configure(tt.params)
			if err == nil {
				t.Fatal("Configure() should return error for invalid params")
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("Error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

func TestParallelProcessorConfigure_PropagatestoWrappedProcessor(t *testing.T) {
	mock := &mockConfigurableProcessor{}
	parallel := NewParallelProcessor(mock)

	params := map[string]string{"key": "value"}
	err := parallel.Configure(params)

	if err != nil {
		t.Fatalf("Configure() returned error: %v", err)
	}

	if !mock.configured {
		t.Error("Configuration was not propagated to wrapped processor")
	}

	if mock.params["key"] != "value" {
		t.Error("Parameters were not passed correctly to wrapped processor")
	}
}

// --- Process Tests - Basic Functionality ---

func TestParallelProcessorProcess_EmptyData(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := []map[string]interface{}{}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Process() returned %d records, expected 0", len(result))
	}
}

func TestParallelProcessorProcess_SmallDatasetSkipsParallel(t *testing.T) {
	var processCalled bool
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			processCalled = true
			return data, nil
		},
	}

	parallel := NewParallelProcessor(processor)
	defer func() { _ = parallel.Close() }()

	// Data smaller than minChunkSize (100)
	ctx := context.Background()
	data := make([]map[string]interface{}, 50)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	if !processCalled {
		t.Error("Small dataset should be processed sequentially")
	}

	if len(result) != 50 {
		t.Errorf("Process() returned %d records, expected 50", len(result))
	}
}

func TestParallelProcessorProcess_LargeDatasetUsesParallel(t *testing.T) {
	var processCount int32
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			atomic.AddInt32(&processCount, 1)
			// Add marker to verify processing
			result := make([]map[string]interface{}, len(data))
			for i, row := range data {
				newRow := make(map[string]interface{})
				for k, v := range row {
					newRow[k] = v
				}
				newRow["processed"] = true
				result[i] = newRow
			}
			return result, nil
		},
	}

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    50,
		MinChunkSize: 100,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	// Large dataset that will be chunked
	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	if len(result) != 200 {
		t.Errorf("Process() returned %d records, expected 200", len(result))
	}

	// Verify parallel processing occurred (multiple process calls)
	count := atomic.LoadInt32(&processCount)
	if count <= 1 {
		t.Errorf("Process was called %d times, expected > 1 for parallel execution", count)
	}

	// Verify all records were processed
	for i, row := range result {
		if !row["processed"].(bool) {
			t.Errorf("Record %d was not processed", i)
		}
	}
}

func TestParallelProcessorProcess_OrderPreserved(t *testing.T) {
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			// Add random delay to cause out-of-order completion
			time.Sleep(time.Duration(len(data)%5) * time.Millisecond)
			return data, nil
		},
	}

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    10,
		MinChunkSize: 5,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 100)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	// Verify order is preserved
	for i, row := range result {
		if row["id"].(int) != i {
			t.Errorf("Record at position %d has id %d, order not preserved", i, row["id"])
		}
	}
}

// --- Context Tests ---

func TestParallelProcessorProcess_ContextCanceled(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)
	defer func() { _ = parallel.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	_, err := parallel.Process(ctx, data)

	if err != context.Canceled {
		t.Errorf("Process() with canceled context returned %v, expected context.Canceled", err)
	}
}

func TestParallelProcessorProcess_ContextTimeout(t *testing.T) {
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return data, nil
		},
	}

	config := ParallelConfig{
		Workers:      2,
		ChunkSize:    50,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	_, err := parallel.Process(ctx, data)

	if err == nil {
		t.Fatal("Process() should timeout")
	}

	// Could be context.DeadlineExceeded or wrapped error
	if err != context.DeadlineExceeded && !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Process() with timeout returned %v, expected timeout error", err)
	}
}

// --- Error Handling Tests ---

func TestParallelProcessorProcess_ProcessorError(t *testing.T) {
	expectedErr := errors.New("processor failed")
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			return nil, expectedErr
		},
	}

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    50,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	_, err := parallel.Process(ctx, data)

	if err == nil {
		t.Fatal("Process() should return error when processor fails")
	}

	// Error should be wrapped
	if !errors.Is(err, expectedErr) {
		t.Errorf("Process() error = %v, should wrap %v", err, expectedErr)
	}
}

func TestParallelProcessorProcess_PartialFailure(t *testing.T) {
	var callCount int32
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			count := atomic.AddInt32(&callCount, 1)
			// Fail on third call
			if count == 3 {
				return nil, errors.New("chunk 3 failed")
			}
			return data, nil
		},
	}

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    50,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	_, err := parallel.Process(ctx, data)

	if err == nil {
		t.Fatal("Process() should return error when a chunk fails")
	}
}

// --- Chunk Size Calculation Tests ---

func TestParallelProcessorCalculateChunkSize(t *testing.T) {
	tests := []struct {
		name         string
		workers      int
		chunkSize    int
		minChunkSize int
		dataSize     int
		expected     int
	}{
		{
			name:         "explicit chunk size",
			workers:      4,
			chunkSize:    500,
			minChunkSize: 100,
			dataSize:     2000,
			expected:     500,
		},
		{
			name:         "auto-calculated",
			workers:      4,
			chunkSize:    0, // Auto
			minChunkSize: 100,
			dataSize:     1000,
			expected:     125, // 1000 / (4 * 2) = 125
		},
		{
			name:         "respects minimum",
			workers:      4,
			chunkSize:    0, // Auto
			minChunkSize: 200,
			dataSize:     400,
			expected:     200, // Auto would be 50, but min is 200
		},
		{
			name:         "at least 1",
			workers:      100,
			chunkSize:    0, // Auto
			minChunkSize: 1,
			dataSize:     10,
			expected:     1, // Ensure at least 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParallelConfig{
				Workers:      tt.workers,
				ChunkSize:    tt.chunkSize,
				MinChunkSize: tt.minChunkSize,
			}
			parallel := NewParallelProcessorWithConfig(&BaseProcessor{}, config)
			defer func() { _ = parallel.Close() }()

			parallel.mu.RLock()
			actual := parallel.calculateChunkSize(tt.dataSize)
			parallel.mu.RUnlock()

			if actual != tt.expected {
				t.Errorf("calculateChunkSize(%d) = %d, expected %d", tt.dataSize, actual, tt.expected)
			}
		})
	}
}

// --- Integration Tests with Wrappers ---

func TestParallelProcessorWithFilterWrapper(t *testing.T) {
	// Create filter that keeps values >= 50
	filter := &mockFilterStrategy{threshold: 50}
	filterWrapper := NewFilterWrapper(filter)

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    25,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(filterWrapper, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 100)
	for i := range data {
		data[i] = map[string]interface{}{"value": i}
	}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	// Should have filtered out values < 50
	expectedCount := 50
	if len(result) != expectedCount {
		t.Errorf("Process() returned %d records, expected %d", len(result), expectedCount)
	}

	// Verify all remaining values >= 50
	for _, row := range result {
		if row["value"].(int) < 50 {
			t.Errorf("Found value %d which should have been filtered", row["value"])
		}
	}
}

func TestParallelProcessorWithValidatorWrapper(t *testing.T) {
	// Validator requires "required_field"
	validator := &mockValidatorStrategy{}
	validatorWrapper := NewValidatorWrapper(validator)

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    25,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(validatorWrapper, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()

	// Valid data
	validData := make([]map[string]interface{}, 100)
	for i := range validData {
		validData[i] = map[string]interface{}{
			"id":             i,
			"required_field": "present",
		}
	}

	result, err := parallel.Process(ctx, validData)

	if err != nil {
		t.Fatalf("Process() with valid data returned error: %v", err)
	}

	if len(result) != 100 {
		t.Errorf("Process() returned %d records, expected 100", len(result))
	}

	// Invalid data
	invalidData := make([]map[string]interface{}, 100)
	for i := range invalidData {
		if i == 50 {
			// Missing required_field
			invalidData[i] = map[string]interface{}{"id": i}
		} else {
			invalidData[i] = map[string]interface{}{
				"id":             i,
				"required_field": "present",
			}
		}
	}

	_, err = parallel.Process(ctx, invalidData)

	if err == nil {
		t.Fatal("Process() should fail with invalid data")
	}
}

func TestParallelProcessorWithTransformWrapper(t *testing.T) {
	// Transformer uppercases names
	transformer := &mockTransformerStrategy{}
	transformWrapper := NewTransformWrapper(transformer)

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    25,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(transformWrapper, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 100)
	for i := range data {
		data[i] = map[string]interface{}{"name": "alice"}
	}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	// Verify transformation occurred
	for i, row := range result {
		name := row["name"].(string)
		if name != "ALICE" {
			t.Errorf("Record %d: name = %q, expected ALICE", i, name)
		}
	}
}

// --- Chaining Tests ---

func TestParallelProcessorWithNext(t *testing.T) {
	processor := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			// Add marker
			result := make([]map[string]interface{}, len(data))
			for i, row := range data {
				newRow := make(map[string]interface{})
				for k, v := range row {
					newRow[k] = v
				}
				newRow["step1"] = true
				result[i] = newRow
			}
			return result, nil
		},
	}

	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    50,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	// Create next processor
	next := &mockProcessorHandler{
		processFunc: func(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
			result := make([]map[string]interface{}, len(data))
			for i, row := range data {
				newRow := make(map[string]interface{})
				for k, v := range row {
					newRow[k] = v
				}
				newRow["step2"] = true
				result[i] = newRow
			}
			return result, nil
		},
	}

	parallel.SetNext(next)

	ctx := context.Background()
	data := make([]map[string]interface{}, 200)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	result, err := parallel.Process(ctx, data)

	if err != nil {
		t.Fatalf("Process() returned error: %v", err)
	}

	// Verify both steps executed
	for i, row := range result {
		if !row["step1"].(bool) {
			t.Errorf("Record %d: step1 not executed", i)
		}
		if !row["step2"].(bool) {
			t.Errorf("Record %d: step2 not executed", i)
		}
	}
}

// --- Close Tests ---

func TestParallelProcessorClose(t *testing.T) {
	var closed bool
	processor := &mockProcessorHandler{
		closeFunc: func() error {
			closed = true
			return nil
		},
	}

	parallel := NewParallelProcessor(processor)

	err := parallel.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	if !closed {
		t.Error("Wrapped processor was not closed")
	}

	// Second close should be safe
	err = parallel.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

func TestParallelProcessorCloseWithContext(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := parallel.CloseWithContext(ctx)
	if err != nil {
		t.Errorf("CloseWithContext() returned error: %v", err)
	}
}

// --- Concurrent Tests ---

func TestParallelProcessorConcurrentAccess(t *testing.T) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)
	defer func() { _ = parallel.Close() }()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			data := make([]map[string]interface{}, 200)
			for j := range data {
				data[j] = map[string]interface{}{"id": j, "goroutine": id}
			}

			_, err := parallel.Process(ctx, data)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent Process() returned error: %v", err)
	}
}

// --- Benchmark Tests ---

func BenchmarkParallelProcessorProcess_Sequential(b *testing.B) {
	processor := &BaseProcessor{}
	parallel := NewParallelProcessor(processor)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 50) // Below minChunkSize
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parallel.Process(ctx, data)
	}
}

func BenchmarkParallelProcessorProcess_Parallel(b *testing.B) {
	processor := &BaseProcessor{}
	config := ParallelConfig{
		Workers:      4,
		ChunkSize:    50,
		MinChunkSize: 10,
	}
	parallel := NewParallelProcessorWithConfig(processor, config)
	defer func() { _ = parallel.Close() }()

	ctx := context.Background()
	data := make([]map[string]interface{}, 1000)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parallel.Process(ctx, data)
	}
}

func BenchmarkParallelProcessorProcess_VariableWorkers(b *testing.B) {
	workerCounts := []int{1, 2, 4, 8, 16}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			processor := &BaseProcessor{}
			config := ParallelConfig{
				Workers:      workers,
				ChunkSize:    100,
				MinChunkSize: 10,
			}
			parallel := NewParallelProcessorWithConfig(processor, config)
			defer func() { _ = parallel.Close() }()

			ctx := context.Background()
			data := make([]map[string]interface{}, 1000)
			for i := range data {
				data[i] = map[string]interface{}{"id": i}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parallel.Process(ctx, data)
			}
		})
	}
}

func BenchmarkParallelProcessorProcess_VariableDataSize(b *testing.B) {
	dataSizes := []int{100, 500, 1000, 5000, 10000}

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			processor := &BaseProcessor{}
			config := ParallelConfig{
				Workers:      4,
				ChunkSize:    100,
				MinChunkSize: 10,
			}
			parallel := NewParallelProcessorWithConfig(processor, config)
			defer func() { _ = parallel.Close() }()

			ctx := context.Background()
			data := make([]map[string]interface{}, size)
			for i := range data {
				data[i] = map[string]interface{}{"id": i}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parallel.Process(ctx, data)
			}
		})
	}
}

// --- Helper Functions ---

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock strategies for integration testing (reuse from wrappers_test.go patterns)

type mockFilterStrategy struct {
	threshold int
}

func (m *mockFilterStrategy) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= m.threshold
	}
	return false
}

type mockValidatorStrategy struct{}

func (m *mockValidatorStrategy) Validate(row map[string]interface{}) error {
	if _, ok := row["required_field"]; !ok {
		return errors.New("missing required_field")
	}
	return nil
}

type mockTransformerStrategy struct{}

func (m *mockTransformerStrategy) Transform(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range row {
		if str, ok := v.(string); ok {
			result[k] = toUpper(str)
		} else {
			result[k] = v
		}
	}
	return result
}

func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
