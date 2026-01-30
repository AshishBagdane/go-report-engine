package processor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- Constructor Tests ---

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name        string
		workers     int
		shouldPanic bool
	}{
		{
			name:        "valid worker count",
			workers:     4,
			shouldPanic: false,
		},
		{
			name:        "single worker",
			workers:     1,
			shouldPanic: false,
		},
		{
			name:        "many workers",
			workers:     100,
			shouldPanic: false,
		},
		{
			name:        "zero workers panics",
			workers:     0,
			shouldPanic: true,
		},
		{
			name:        "negative workers panics",
			workers:     -1,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.shouldPanic && r == nil {
					t.Error("Expected panic but got none")
				}
				if !tt.shouldPanic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			pool := NewWorkerPool(tt.workers)
			if !tt.shouldPanic {
				if pool == nil {
					t.Error("NewWorkerPool returned nil")
				}
				if pool.Workers() != tt.workers {
					t.Errorf("Workers() = %d, expected %d", pool.Workers(), tt.workers)
				}
			}
		})
	}
}

// --- Basic Functionality Tests ---

func TestWorkerPoolProcessChunks_Empty(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := []WorkChunk{}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	results, err := pool.ProcessChunks(ctx, chunks, task)

	if err != nil {
		t.Errorf("ProcessChunks() with empty chunks returned error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("ProcessChunks() returned %d results, expected 0", len(results))
	}
}

func TestWorkerPoolProcessChunks_SingleChunk(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := []WorkChunk{
		{
			Data:  []map[string]interface{}{{"id": 1}, {"id": 2}},
			Index: 0,
			Start: 0,
			End:   2,
		},
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	results, err := pool.ProcessChunks(ctx, chunks, task)

	if err != nil {
		t.Fatalf("ProcessChunks() returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("ProcessChunks() returned %d results, expected 1", len(results))
	}

	if len(results[0].Data) != 2 {
		t.Errorf("Result data has %d records, expected 2", len(results[0].Data))
	}
}

func TestWorkerPoolProcessChunks_MultipleChunks(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := []WorkChunk{
		{Data: []map[string]interface{}{{"id": 1}}, Index: 0, Start: 0, End: 1},
		{Data: []map[string]interface{}{{"id": 2}}, Index: 1, Start: 1, End: 2},
		{Data: []map[string]interface{}{{"id": 3}}, Index: 2, Start: 2, End: 3},
		{Data: []map[string]interface{}{{"id": 4}}, Index: 3, Start: 3, End: 4},
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Add marker to verify processing occurred
		result := make([]map[string]interface{}, len(chunk.Data))
		for i, row := range chunk.Data {
			newRow := make(map[string]interface{})
			for k, v := range row {
				newRow[k] = v
			}
			newRow["processed"] = true
			result[i] = newRow
		}
		return result, nil
	}

	results, err := pool.ProcessChunks(ctx, chunks, task)

	if err != nil {
		t.Fatalf("ProcessChunks() returned error: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("ProcessChunks() returned %d results, expected 4", len(results))
	}

	// Verify results are in order
	for i, result := range results {
		if result.Index != i {
			t.Errorf("Result %d has index %d, expected %d", i, result.Index, i)
		}
		if !result.Data[0]["processed"].(bool) {
			t.Errorf("Result %d was not processed", i)
		}
	}
}

// --- Context Cancellation Tests ---

func TestWorkerPoolProcessChunks_ContextCanceled(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	chunks := []WorkChunk{
		{Data: []map[string]interface{}{{"id": 1}}, Index: 0},
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	_, err := pool.ProcessChunks(ctx, chunks, task)

	if err != context.Canceled {
		t.Errorf("ProcessChunks() with canceled context returned %v, expected context.Canceled", err)
	}
}

func TestWorkerPoolProcessChunks_ContextTimeout(t *testing.T) {
	pool := NewWorkerPool(2)
	defer func() { _ = pool.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Create many chunks to ensure timeout
	chunks := make([]WorkChunk, 100)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Simulate slow processing
		time.Sleep(100 * time.Millisecond)
		return chunk.Data, nil
	}

	_, err := pool.ProcessChunks(ctx, chunks, task)

	if err != context.DeadlineExceeded {
		t.Errorf("ProcessChunks() with timeout returned %v, expected context.DeadlineExceeded", err)
	}
}

func TestWorkerPoolProcessChunks_ContextCanceledDuringProcessing(t *testing.T) {
	pool := NewWorkerPool(2)
	defer func() { _ = pool.Close() }()

	ctx, cancel := context.WithCancel(context.Background())

	chunks := make([]WorkChunk, 10)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
		}
	}

	var processedCount int32

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Cancel after first few chunks
		count := atomic.AddInt32(&processedCount, 1)
		if count == 3 {
			cancel()
		}

		// Check context
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		time.Sleep(10 * time.Millisecond)
		return chunk.Data, nil
	}

	_, err := pool.ProcessChunks(ctx, chunks, task)

	if err != context.Canceled {
		t.Errorf("ProcessChunks() returned %v, expected context.Canceled", err)
	}
}

// --- Error Handling Tests ---

func TestWorkerPoolProcessChunks_TaskError(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := []WorkChunk{
		{Data: []map[string]interface{}{{"id": 1}}, Index: 0},
		{Data: []map[string]interface{}{{"id": 2}}, Index: 1},
		{Data: []map[string]interface{}{{"id": 3}}, Index: 2},
	}

	expectedErr := errors.New("task failed")

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Fail on second chunk
		if chunk.Index == 1 {
			return nil, expectedErr
		}
		return chunk.Data, nil
	}

	_, err := pool.ProcessChunks(ctx, chunks, task)

	if err == nil {
		t.Fatal("ProcessChunks() should return error when task fails")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("ProcessChunks() error = %v, expected %v", err, expectedErr)
	}
}

func TestWorkerPoolProcessChunks_MultipleErrors(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := make([]WorkChunk, 10)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Multiple chunks fail
		if chunk.Index%2 == 0 {
			return nil, fmt.Errorf("chunk %d failed", chunk.Index)
		}
		return chunk.Data, nil
	}

	_, err := pool.ProcessChunks(ctx, chunks, task)

	if err == nil {
		t.Fatal("ProcessChunks() should return error when tasks fail")
	}

	// Should return the first error encountered
	// The exact error depends on execution order, so just verify we got an error
}

// --- Ordering Tests ---

func TestWorkerPoolProcessChunks_OrderPreserved(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()

	// Create chunks in order
	numChunks := 20
	chunks := make([]WorkChunk, numChunks)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
			Start: i,
			End:   i + 1,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Add random delay to cause out-of-order completion
		delay := time.Duration(chunk.Index%5) * time.Millisecond
		time.Sleep(delay)
		return chunk.Data, nil
	}

	results, err := pool.ProcessChunks(ctx, chunks, task)

	if err != nil {
		t.Fatalf("ProcessChunks() returned error: %v", err)
	}

	// Verify results are in original order
	for i, result := range results {
		if result.Index != i {
			t.Errorf("Result at position %d has index %d, order not preserved", i, result.Index)
		}
	}
}

// --- Worker Count Tests ---

func TestWorkerPoolProcessChunks_MoreWorkersThanChunks(t *testing.T) {
	pool := NewWorkerPool(10)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := []WorkChunk{
		{Data: []map[string]interface{}{{"id": 1}}, Index: 0},
		{Data: []map[string]interface{}{{"id": 2}}, Index: 1},
	}

	var activeWorkers int32

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		count := atomic.AddInt32(&activeWorkers, 1)
		defer atomic.AddInt32(&activeWorkers, -1)

		// Verify we don't have more active workers than chunks
		if count > 2 {
			t.Errorf("More active workers (%d) than chunks (2)", count)
		}

		time.Sleep(10 * time.Millisecond)
		return chunk.Data, nil
	}

	results, err := pool.ProcessChunks(ctx, chunks, task)

	if err != nil {
		t.Fatalf("ProcessChunks() returned error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("ProcessChunks() returned %d results, expected 2", len(results))
	}
}

func TestWorkerPoolProcessChunks_FewerWorkersThanChunks(t *testing.T) {
	pool := NewWorkerPool(2)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := make([]WorkChunk, 10)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
		}
	}

	var maxActiveWorkers int32
	var currentActive int32

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		active := atomic.AddInt32(&currentActive, 1)
		defer atomic.AddInt32(&currentActive, -1)

		// Track max concurrent workers
		for {
			max := atomic.LoadInt32(&maxActiveWorkers)
			if active <= max || atomic.CompareAndSwapInt32(&maxActiveWorkers, max, active) {
				break
			}
		}

		time.Sleep(5 * time.Millisecond)
		return chunk.Data, nil
	}

	results, err := pool.ProcessChunks(ctx, chunks, task)

	if err != nil {
		t.Fatalf("ProcessChunks() returned error: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("ProcessChunks() returned %d results, expected 10", len(results))
	}

	// Verify we didn't exceed worker limit
	max := atomic.LoadInt32(&maxActiveWorkers)
	if max > 2 {
		t.Errorf("Max active workers was %d, expected <= 2", max)
	}
}

// --- Close Tests ---

func TestWorkerPoolClose(t *testing.T) {
	pool := NewWorkerPool(4)

	if pool.IsClosed() {
		t.Error("Newly created pool should not be closed")
	}

	err := pool.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	if !pool.IsClosed() {
		t.Error("Pool should be closed after Close()")
	}

	// Close again should be safe (idempotent)
	err = pool.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

func TestWorkerPoolClose_RejectsNewWork(t *testing.T) {
	pool := NewWorkerPool(4)
	pool.Close()

	ctx := context.Background()
	chunks := []WorkChunk{
		{Data: []map[string]interface{}{{"id": 1}}, Index: 0},
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	_, err := pool.ProcessChunks(ctx, chunks, task)

	if err == nil {
		t.Error("ProcessChunks() should fail on closed pool")
	}

	if err.Error() != "worker pool is closed" {
		t.Errorf("ProcessChunks() error = %v, expected 'worker pool is closed'", err)
	}
}

func TestWorkerPoolCloseWithContext(t *testing.T) {
	pool := NewWorkerPool(4)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pool.CloseWithContext(ctx)
	if err != nil {
		t.Errorf("CloseWithContext() returned error: %v", err)
	}

	if !pool.IsClosed() {
		t.Error("Pool should be closed after CloseWithContext()")
	}
}

func TestWorkerPoolCloseWithContext_Timeout(t *testing.T) {
	pool := NewWorkerPool(2)

	// CloseWithContext should complete quickly since it just marks pool as closed
	// (it no longer waits for active work to complete)
	closeCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := pool.CloseWithContext(closeCtx)

	// Should succeed without timeout since Close is now fast
	if err != nil {
		t.Errorf("CloseWithContext() returned error: %v", err)
	}

	// Pool should be marked as closed
	if !pool.IsClosed() {
		t.Error("Pool should be marked closed")
	}

	// Verify closed pool rejects new work
	ctx := context.Background()
	chunks := []WorkChunk{
		{Data: []map[string]interface{}{{"id": 1}}, Index: 0},
	}
	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	_, err = pool.ProcessChunks(ctx, chunks, task)
	if err == nil {
		t.Error("Closed pool should reject new work")
	}
}

// --- Concurrent Tests ---

func TestWorkerPoolProcessChunks_Concurrent(t *testing.T) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			chunks := []WorkChunk{
				{Data: []map[string]interface{}{{"id": id}}, Index: 0},
			}

			task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
				return chunk.Data, nil
			}

			_, err := pool.ProcessChunks(ctx, chunks, task)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent ProcessChunks() returned error: %v", err)
	}
}

// --- sortWorkResults Tests ---

func TestSortWorkResults(t *testing.T) {
	tests := []struct {
		name     string
		input    []WorkResult
		expected []int // Expected order of indices
	}{
		{
			name: "already sorted",
			input: []WorkResult{
				{Index: 0}, {Index: 1}, {Index: 2},
			},
			expected: []int{0, 1, 2},
		},
		{
			name: "reverse order",
			input: []WorkResult{
				{Index: 2}, {Index: 1}, {Index: 0},
			},
			expected: []int{0, 1, 2},
		},
		{
			name: "random order",
			input: []WorkResult{
				{Index: 3}, {Index: 1}, {Index: 4}, {Index: 0}, {Index: 2},
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:     "empty",
			input:    []WorkResult{},
			expected: []int{},
		},
		{
			name: "single element",
			input: []WorkResult{
				{Index: 0},
			},
			expected: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortWorkResults(tt.input)

			if len(tt.input) != len(tt.expected) {
				t.Fatalf("Result length = %d, expected %d", len(tt.input), len(tt.expected))
			}

			for i, expectedIndex := range tt.expected {
				if tt.input[i].Index != expectedIndex {
					t.Errorf("Position %d: index = %d, expected %d", i, tt.input[i].Index, expectedIndex)
				}
			}
		})
	}
}

// --- Benchmark Tests ---

func BenchmarkWorkerPoolProcessChunks_SmallChunks(b *testing.B) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := make([]WorkChunk, 10)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pool.ProcessChunks(ctx, chunks, task)
	}
}

func BenchmarkWorkerPoolProcessChunks_LargeChunks(b *testing.B) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := make([]WorkChunk, 10)
	for i := range chunks {
		data := make([]map[string]interface{}, 1000)
		for j := range data {
			data[j] = map[string]interface{}{"id": j}
		}
		chunks[i] = WorkChunk{
			Data:  data,
			Index: i,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pool.ProcessChunks(ctx, chunks, task)
	}
}

func BenchmarkWorkerPoolProcessChunks_ManyChunks(b *testing.B) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := make([]WorkChunk, 100)
	for i := range chunks {
		chunks[i] = WorkChunk{
			Data:  []map[string]interface{}{{"id": i}},
			Index: i,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		return chunk.Data, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pool.ProcessChunks(ctx, chunks, task)
	}
}

func BenchmarkWorkerPoolProcessChunks_CPUBound(b *testing.B) {
	pool := NewWorkerPool(4)
	defer func() { _ = pool.Close() }()

	ctx := context.Background()
	chunks := make([]WorkChunk, 20)
	for i := range chunks {
		data := make([]map[string]interface{}, 100)
		for j := range data {
			data[j] = map[string]interface{}{"value": j}
		}
		chunks[i] = WorkChunk{
			Data:  data,
			Index: i,
		}
	}

	task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
		// Simulate CPU-bound work
		result := make([]map[string]interface{}, len(chunk.Data))
		for i, row := range chunk.Data {
			newRow := make(map[string]interface{})
			for k, v := range row {
				newRow[k] = v
			}
			// Some computation
			if val, ok := row["value"].(int); ok {
				newRow["squared"] = val * val
			}
			result[i] = newRow
		}
		return result, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pool.ProcessChunks(ctx, chunks, task)
	}
}

func BenchmarkWorkerPoolProcessChunks_VariableWorkers(b *testing.B) {
	workerCounts := []int{1, 2, 4, 8, 16}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			pool := NewWorkerPool(workers)
			defer func() { _ = pool.Close() }()

			ctx := context.Background()
			chunks := make([]WorkChunk, 50)
			for i := range chunks {
				chunks[i] = WorkChunk{
					Data:  []map[string]interface{}{{"id": i}},
					Index: i,
				}
			}

			task := func(ctx context.Context, chunk WorkChunk) ([]map[string]interface{}, error) {
				// Minimal work
				return chunk.Data, nil
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				pool.ProcessChunks(ctx, chunks, task)
			}
		})
	}
}
