package registry

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// mockProcessor is a simple test implementation of ProcessorHandler
type mockProcessor struct {
	name string
	processor.BaseProcessor
}

func (m *mockProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Simply pass through
	return m.BaseProcessor.Process(ctx, data)
}

// mockProcessorFactory creates a new mock processor
func mockProcessorFactory(name string) ProcessorFactory {
	return func() processor.ProcessorHandler {
		return &mockProcessor{name: name}
	}
}

// mockFilterStrategy for testing type-safe registration
type mockFilterStrategy struct {
	threshold int
}

func (m *mockFilterStrategy) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= m.threshold
	}
	return false
}

// mockValidatorStrategy for testing type-safe registration
type mockValidatorStrategy struct{}

func (m *mockValidatorStrategy) Validate(row map[string]interface{}) error {
	if _, ok := row["required_field"]; !ok {
		return fmt.Errorf("missing required field")
	}
	return nil
}

// mockTransformerStrategy for testing type-safe registration
type mockTransformerStrategy struct {
	multiplier int
}

func (m *mockTransformerStrategy) Transform(row map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range row {
		if val, ok := v.(int); ok {
			result[k] = val * m.multiplier
		} else {
			result[k] = v
		}
	}
	return result
}

// TestRegisterProcessor tests basic registration functionality
func TestRegisterProcessor(t *testing.T) {
	// Clean state for test
	ClearProcessors()

	tests := []struct {
		name          string
		processorName string
		factory       ProcessorFactory
		shouldPanic   bool
	}{
		{
			name:          "valid registration",
			processorName: "test_processor",
			factory:       mockProcessorFactory("test"),
			shouldPanic:   false,
		},
		{
			name:          "empty name panics",
			processorName: "",
			factory:       mockProcessorFactory("test"),
			shouldPanic:   true,
		},
		{
			name:          "nil factory panics",
			processorName: "test_nil",
			factory:       nil,
			shouldPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterProcessor did not panic for %s", tt.name)
					}
				}()
			}

			RegisterProcessor(tt.processorName, tt.factory)

			if !tt.shouldPanic {
				if !IsProcessorRegistered(tt.processorName) {
					t.Errorf("Processor %s was not registered", tt.processorName)
				}
			}
		})
	}
}

// TestGetProcessor tests retrieval of registered processors
func TestGetProcessor(t *testing.T) {
	ClearProcessors()

	// Register test processor
	RegisterProcessor("test_processor", mockProcessorFactory("test"))

	tests := []struct {
		name        string
		lookup      string
		shouldError bool
	}{
		{
			name:        "get existing processor",
			lookup:      "test_processor",
			shouldError: false,
		},
		{
			name:        "get non-existent processor",
			lookup:      "nonexistent",
			shouldError: true,
		},
		{
			name:        "empty name returns error",
			lookup:      "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetProcessor(tt.lookup)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
				if result == nil {
					t.Errorf("Expected processor instance, got nil")
				}
			}
		})
	}
}

// TestRegisterFilter tests type-safe filter registration
func TestRegisterFilter(t *testing.T) {
	ClearProcessors()

	filter := &mockFilterStrategy{threshold: 50}
	RegisterFilter("test_filter", filter)

	if !IsProcessorRegistered("test_filter") {
		t.Error("Filter was not registered as processor")
	}

	// Verify we can get it back
	proc, err := GetProcessor("test_filter")
	if err != nil {
		t.Fatalf("Failed to get registered filter: %v", err)
	}

	// Test that it works
	testData := []map[string]interface{}{
		{"id": 1, "value": 30},
		{"id": 2, "value": 60},
	}

	ctx := context.Background()
	result, err := proc.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Filter process failed: %v", err)
	}

	// Should filter out value < 50
	if len(result) != 1 {
		t.Errorf("Filter returned %d records, expected 1", len(result))
	}
}

// TestRegisterValidator tests type-safe validator registration
func TestRegisterValidator(t *testing.T) {
	ClearProcessors()

	validator := &mockValidatorStrategy{}
	RegisterValidator("test_validator", validator)

	if !IsProcessorRegistered("test_validator") {
		t.Error("Validator was not registered as processor")
	}

	// Verify we can get it back
	proc, err := GetProcessor("test_validator")
	if err != nil {
		t.Fatalf("Failed to get registered validator: %v", err)
	}

	// Test that it works - valid data
	validData := []map[string]interface{}{
		{"id": 1, "required_field": "present"},
	}

	ctx := context.Background()
	result, err := proc.Process(ctx, validData)
	if err != nil {
		t.Errorf("Validator should pass valid data: %v", err)
	}
	if len(result) != 1 {
		t.Error("Validator should return valid data")
	}

	// Test that it works - invalid data
	invalidData := []map[string]interface{}{
		{"id": 1}, // Missing required_field
	}

	_, err = proc.Process(ctx, invalidData)
	if err == nil {
		t.Error("Validator should reject invalid data")
	}
}

// TestRegisterTransformer tests type-safe transformer registration
func TestRegisterTransformer(t *testing.T) {
	ClearProcessors()

	transformer := &mockTransformerStrategy{multiplier: 2}
	RegisterTransformer("test_transformer", transformer)

	if !IsProcessorRegistered("test_transformer") {
		t.Error("Transformer was not registered as processor")
	}

	// Verify we can get it back
	proc, err := GetProcessor("test_transformer")
	if err != nil {
		t.Fatalf("Failed to get registered transformer: %v", err)
	}

	// Test that it works
	testData := []map[string]interface{}{
		{"id": 1, "value": 10},
	}

	ctx := context.Background()
	result, err := proc.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Transformer process failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result))
	}

	// Value should be multiplied by 2
	if result[0]["value"].(int) != 20 {
		t.Errorf("Transformer did not multiply value correctly, got %v", result[0]["value"])
	}
}

// TestListProcessors tests listing of registered processors
func TestListProcessors(t *testing.T) {
	ClearProcessors()

	// Empty list
	list := ListProcessors()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %v", list)
	}

	// Add processors
	RegisterProcessor("proc1", mockProcessorFactory("1"))
	RegisterFilter("filter1", &mockFilterStrategy{threshold: 50})
	RegisterValidator("validator1", &mockValidatorStrategy{})

	list = ListProcessors()

	if len(list) != 3 {
		t.Errorf("Expected 3 processors, got %d", len(list))
	}

	// Verify sorted order
	expected := []string{"filter1", "proc1", "validator1"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("Expected processors[%d] = %s, got %s", i, name, list[i])
		}
	}
}

// TestIsProcessorRegistered tests checking processor existence
func TestIsProcessorRegistered(t *testing.T) {
	ClearProcessors()

	RegisterProcessor("existing", mockProcessorFactory("test"))

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"existing processor", "existing", true},
		{"non-existent processor", "nonexistent", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsProcessorRegistered(tt.lookup)
			if result != tt.expected {
				t.Errorf("IsProcessorRegistered(%s) = %v, expected %v", tt.lookup, result, tt.expected)
			}
		})
	}
}

// TestUnregisterProcessor tests removal of processors
func TestUnregisterProcessor(t *testing.T) {
	ClearProcessors()

	RegisterProcessor("to_remove", mockProcessorFactory("test"))

	if !IsProcessorRegistered("to_remove") {
		t.Fatal("Processor should be registered before removal")
	}

	UnregisterProcessor("to_remove")

	if IsProcessorRegistered("to_remove") {
		t.Error("Processor should not be registered after removal")
	}

	// Unregistering non-existent processor should not panic
	UnregisterProcessor("nonexistent")
}

// TestClearProcessors tests clearing all processors
func TestClearProcessors(t *testing.T) {
	ClearProcessors()

	RegisterProcessor("p1", mockProcessorFactory("1"))
	RegisterFilter("f1", &mockFilterStrategy{threshold: 50})
	RegisterValidator("v1", &mockValidatorStrategy{})

	if ProcessorCount() != 3 {
		t.Fatalf("Expected 3 processors, got %d", ProcessorCount())
	}

	ClearProcessors()

	if ProcessorCount() != 0 {
		t.Errorf("Expected 0 processors after clear, got %d", ProcessorCount())
	}
}

// TestProcessorCount tests counting registered processors
func TestProcessorCount(t *testing.T) {
	ClearProcessors()

	if ProcessorCount() != 0 {
		t.Errorf("Expected 0 processors initially, got %d", ProcessorCount())
	}

	RegisterProcessor("p1", mockProcessorFactory("1"))
	if ProcessorCount() != 1 {
		t.Errorf("Expected 1 processor, got %d", ProcessorCount())
	}

	RegisterFilter("f1", &mockFilterStrategy{threshold: 50})
	if ProcessorCount() != 2 {
		t.Errorf("Expected 2 processors, got %d", ProcessorCount())
	}

	UnregisterProcessor("p1")
	if ProcessorCount() != 1 {
		t.Errorf("Expected 1 processor after removal, got %d", ProcessorCount())
	}
}

// configurableFilter for testing configuration
type configurableFilter struct {
	threshold  int
	configured bool
}

func (c *configurableFilter) Keep(row map[string]interface{}) bool {
	if val, ok := row["value"].(int); ok {
		return val >= c.threshold
	}
	return false
}

func (c *configurableFilter) Configure(params map[string]string) error {
	c.configured = true
	thresholdStr, ok := params["threshold"]
	if !ok {
		return api.ErrMissingParam("threshold")
	}
	// Simple string to int conversion for test
	if thresholdStr == "100" {
		c.threshold = 100
	} else {
		c.threshold = 50
	}
	return nil
}

// TestRegisterFilterWithConfiguration tests configurable filter
func TestRegisterFilterWithConfiguration(t *testing.T) {
	ClearProcessors()

	filter := &configurableFilter{}
	RegisterFilter("configurable_filter", filter)

	proc, err := GetProcessor("configurable_filter")
	if err != nil {
		t.Fatalf("Failed to get processor: %v", err)
	}

	// Configure it
	if configurable, ok := proc.(api.Configurable); ok {
		err := configurable.Configure(map[string]string{"threshold": "100"})
		if err != nil {
			t.Fatalf("Configuration failed: %v", err)
		}
	} else {
		t.Fatal("Processor should be configurable")
	}

	// Test it works with configured threshold
	testData := []map[string]interface{}{
		{"id": 1, "value": 75},
		{"id": 2, "value": 150},
	}

	ctx := context.Background()
	result, err := proc.Process(ctx, testData)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// With threshold=100, should only keep value >= 100
	if len(result) != 1 {
		t.Errorf("Expected 1 record with threshold 100, got %d", len(result))
	}
}

// TestProcessorConcurrentRegister tests concurrent registration
func TestProcessorConcurrentRegister(t *testing.T) {
	ClearProcessors()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("proc_%d", id%26)
			RegisterProcessor(name, mockProcessorFactory(name))
		}(i)
	}

	wg.Wait()

	// Should have registered processors
	count := ProcessorCount()
	if count == 0 || count > 26 {
		t.Errorf("Expected between 1 and 26 processors, got %d", count)
	}
}

// TestProcessorConcurrentGet tests concurrent retrieval
func TestProcessorConcurrentGet(t *testing.T) {
	ClearProcessors()

	RegisterProcessor("test", mockProcessorFactory("test"))

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := GetProcessor("test")
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent get failed: %v", err)
	}
}

// TestProcessorConcurrentMixedOperations tests all operations concurrently
func TestProcessorConcurrentMixedOperations(t *testing.T) {
	ClearProcessors()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 5) // 5 different operations

	// Register
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("proc_%d", id%5)
			RegisterProcessor(name, mockProcessorFactory(name))
		}(i)
	}

	// Get
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("proc_%d", id%5)
			GetProcessor(name)
		}(i)
	}

	// List
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ListProcessors()
		}()
	}

	// IsRegistered
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("proc_%d", id%5)
			IsProcessorRegistered(name)
		}(i)
	}

	// Count
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ProcessorCount()
		}()
	}

	wg.Wait()

	// If we get here without deadlock or panic, test passes
	t.Logf("Successfully completed %d concurrent operations", goroutines*5)
}

// TestProcessorRaceDetector should be run with -race flag
func TestProcessorRaceDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detector test in short mode")
	}

	ClearProcessors()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			RegisterProcessor("test", mockProcessorFactory("test"))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			GetProcessor("test")
			ListProcessors()
			IsProcessorRegistered("test")
			ProcessorCount()
		}
		done <- true
	}()

	<-done
	<-done
}

// BenchmarkRegisterProcessor benchmarks processor registration
func BenchmarkRegisterProcessor(b *testing.B) {
	ClearProcessors()
	factory := mockProcessorFactory("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterProcessor("bench_processor", factory)
	}
}

// BenchmarkGetProcessor benchmarks processor retrieval
func BenchmarkGetProcessor(b *testing.B) {
	ClearProcessors()
	RegisterProcessor("bench", mockProcessorFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetProcessor("bench")
	}
}

// BenchmarkGetProcessorParallel benchmarks concurrent processor retrieval
func BenchmarkGetProcessorParallel(b *testing.B) {
	ClearProcessors()
	RegisterProcessor("bench", mockProcessorFactory("test"))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetProcessor("bench")
		}
	})
}

// BenchmarkRegisterFilter benchmarks filter registration
func BenchmarkRegisterFilter(b *testing.B) {
	ClearProcessors()
	filter := &mockFilterStrategy{threshold: 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterFilter("bench_filter", filter)
	}
}

// BenchmarkListProcessors benchmarks listing processors
func BenchmarkListProcessors(b *testing.B) {
	ClearProcessors()
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("proc_%d", i)
		RegisterProcessor(name, mockProcessorFactory(name))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListProcessors()
	}
}
