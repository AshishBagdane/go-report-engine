package processor

import (
	"context"
	"fmt"
	"testing"
)

// TestBaseProcessorZeroValue tests the zero value behavior
func TestBaseProcessorZeroValue(t *testing.T) {
	var proc BaseProcessor

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	ctx := context.Background()
	result, err := proc.Process(ctx, testData)

	if err != nil {
		t.Errorf("Process() returned unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("Process() returned nil")
	}

	// Should return data as-is when no next processor
	if len(result) != len(testData) {
		t.Errorf("Process() returned %d records, expected %d", len(result), len(testData))
	}
}

// TestBaseProcessorSetNext tests setting the next processor
func TestBaseProcessorSetNext(t *testing.T) {
	proc1 := &BaseProcessor{}
	proc2 := &BaseProcessor{}

	proc1.SetNext(proc2)

	// Verify next was set (we can't check directly, but Process should work)
	testData := []map[string]interface{}{
		{"id": 1},
	}

	ctx := context.Background()
	result, err := proc1.Process(ctx, testData)
	if err != nil {
		t.Errorf("Process() with next processor failed: %v", err)
	}

	if result == nil {
		t.Error("Process() returned nil")
	}
}

// TestBaseProcessorProcessPassthrough tests data passthrough
func TestBaseProcessorProcessPassthrough(t *testing.T) {
	proc := &BaseProcessor{}

	tests := []struct {
		name     string
		input    []map[string]interface{}
		expected int
	}{
		{
			name:     "empty data",
			input:    []map[string]interface{}{},
			expected: 0,
		},
		{
			name: "single record",
			input: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
			},
			expected: 1,
		},
		{
			name: "multiple records",
			input: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
				{"id": 3, "name": "Charlie"},
			},
			expected: 3,
		},
		{
			name: "records with different fields",
			input: []map[string]interface{}{
				{"id": 1, "type": "user"},
				{"name": "test", "value": 42},
			},
			expected: 2,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := proc.Process(ctx, tt.input)

			if err != nil {
				t.Errorf("Process() returned error: %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("Process() returned %d records, expected %d", len(result), tt.expected)
			}

			// Verify data is unmodified
			for i := range result {
				if i >= len(tt.input) {
					break
				}
				for key, val := range tt.input[i] {
					if result[i][key] != val {
						t.Errorf("Record %d: field %s was modified", i, key)
					}
				}
			}
		})
	}
}

// TestBaseProcessorChaining tests chaining multiple processors
func TestBaseProcessorChaining(t *testing.T) {
	proc1 := &BaseProcessor{}
	proc2 := &BaseProcessor{}
	proc3 := &BaseProcessor{}

	// Chain: proc1 -> proc2 -> proc3
	proc1.SetNext(proc2)
	proc2.SetNext(proc3)

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	ctx := context.Background()
	result, err := proc1.Process(ctx, testData)

	if err != nil {
		t.Errorf("Chained Process() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Chained Process() returned nil")
	}

	if len(result) != len(testData) {
		t.Errorf("Chained Process() returned %d records, expected %d", len(result), len(testData))
	}
}

// mockFilterProcessor for testing custom processors in chain
type mockFilterProcessor struct {
	BaseProcessor
	keepCondition func(map[string]interface{}) bool
}

func (m *mockFilterProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	var filtered []map[string]interface{}
	for _, row := range data {
		if m.keepCondition(row) {
			filtered = append(filtered, row)
		}
	}
	return m.BaseProcessor.Process(ctx, filtered)
}

// TestBaseProcessorWithCustomProcessor tests chaining with custom processor
func TestBaseProcessorWithCustomProcessor(t *testing.T) {
	// Create a filter that keeps only records with id > 1
	filter := &mockFilterProcessor{
		keepCondition: func(row map[string]interface{}) bool {
			if id, ok := row["id"].(int); ok {
				return id > 1
			}
			return false
		},
	}

	base := &BaseProcessor{}
	filter.SetNext(base)

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
	}

	ctx := context.Background()
	result, err := filter.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() failed: %v", err)
	}

	expectedCount := 2
	if len(result) != expectedCount {
		t.Errorf("Process() returned %d records, expected %d", len(result), expectedCount)
	}

	// Verify filtered records
	for _, row := range result {
		if id := row["id"].(int); id <= 1 {
			t.Errorf("Found record with id %d, should have been filtered", id)
		}
	}
}

// mockErrorProcessor for testing error propagation
type mockErrorProcessor struct {
	BaseProcessor
	shouldError bool
	errorMsg    string
}

func (m *mockErrorProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return m.BaseProcessor.Process(ctx, data)
}

// TestBaseProcessorErrorPropagation tests error handling in chain
func TestBaseProcessorErrorPropagation(t *testing.T) {
	errorProc := &mockErrorProcessor{
		shouldError: true,
		errorMsg:    "test error",
	}

	base := &BaseProcessor{}
	errorProc.SetNext(base)

	testData := []map[string]interface{}{
		{"id": 1},
	}

	ctx := context.Background()
	result, err := errorProc.Process(ctx, testData)

	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}

	expectedMsg := "test error"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, expected %q", err.Error(), expectedMsg)
	}
}

// TestBaseProcessorNilData tests handling of nil data
func TestBaseProcessorNilData(t *testing.T) {
	proc := &BaseProcessor{}

	ctx := context.Background()
	result, err := proc.Process(ctx, nil)

	if err != nil {
		t.Errorf("Process(nil) returned error: %v", err)
	}

	if result != nil {
		t.Error("Process(nil) should return nil")
	}
}

// TestBaseProcessorContextCancellation tests context cancellation
func TestBaseProcessorContextCancellation(t *testing.T) {
	proc := &BaseProcessor{}

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := proc.Process(ctx, testData)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// TestBaseProcessorImplementsInterface verifies interface implementation
func TestBaseProcessorImplementsInterface(t *testing.T) {
	var _ ProcessorHandler = (*BaseProcessor)(nil)
}

// TestBaseProcessorLargeDataset tests with large dataset
func TestBaseProcessorLargeDataset(t *testing.T) {
	proc := &BaseProcessor{}

	// Create large dataset
	const recordCount = 10000
	testData := make([]map[string]interface{}, recordCount)
	for i := 0; i < recordCount; i++ {
		testData[i] = map[string]interface{}{
			"id":    i,
			"value": i * 2,
		}
	}

	ctx := context.Background()
	result, err := proc.Process(ctx, testData)

	if err != nil {
		t.Fatalf("Process() failed on large dataset: %v", err)
	}

	if len(result) != recordCount {
		t.Errorf("Process() returned %d records, expected %d", len(result), recordCount)
	}
}

// TestBaseProcessorConcurrentAccess tests concurrent access
func TestBaseProcessorConcurrentAccess(t *testing.T) {
	proc := &BaseProcessor{}

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	const goroutines = 10
	errors := make(chan error, goroutines)
	results := make(chan []map[string]interface{}, goroutines)

	ctx := context.Background()

	for i := 0; i < goroutines; i++ {
		go func() {
			result, err := proc.Process(ctx, testData)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent Process() failed: %v", err)
		case result := <-results:
			if len(result) != len(testData) {
				t.Errorf("Concurrent Process() returned wrong count")
			}
		}
	}
}

// BenchmarkBaseProcessorProcess benchmarks basic processing
func BenchmarkBaseProcessorProcess(b *testing.B) {
	proc := &BaseProcessor{}
	testData := []map[string]interface{}{
		{"id": 1, "name": "test", "value": 42},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.Process(ctx, testData)
	}
}

// BenchmarkBaseProcessorProcessLarge benchmarks large dataset
func BenchmarkBaseProcessorProcessLarge(b *testing.B) {
	proc := &BaseProcessor{}

	testData := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		testData[i] = map[string]interface{}{
			"id":    i,
			"value": i * 2,
		}
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc.Process(ctx, testData)
	}
}

// BenchmarkBaseProcessorChain benchmarks chained processors
func BenchmarkBaseProcessorChain(b *testing.B) {
	proc1 := &BaseProcessor{}
	proc2 := &BaseProcessor{}
	proc3 := &BaseProcessor{}

	proc1.SetNext(proc2)
	proc2.SetNext(proc3)

	testData := []map[string]interface{}{
		{"id": 1, "name": "test"},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proc1.Process(ctx, testData)
	}
}
