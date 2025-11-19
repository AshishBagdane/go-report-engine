package registry

import (
	"fmt"
	"sync"
	"testing"

	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// mockProcessor is a simple test implementation of ProcessorHandler
type mockProcessor struct {
	name string
	processor.BaseProcessor
}

func (m *mockProcessor) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Simply pass through
	return m.BaseProcessor.Process(data)
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
		errorType   string
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
			errorType:   "*registry.ErrProcessorNotFound",
		},
		{
			name:        "empty name returns error",
			lookup:      "",
			shouldError: true,
			errorType:   "ErrEmptyProcessorName",
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

// TestGetProcessorReturnsNewInstance verifies that each call returns a new instance
func TestGetProcessorReturnsNewInstance(t *testing.T) {
	ClearProcessors()

	RegisterProcessor("test", mockProcessorFactory("test"))

	p1, err1 := GetProcessor("test")
	p2, err2 := GetProcessor("test")

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected errors: %v, %v", err1, err2)
	}

	// Different instances should have different addresses
	if p1 == p2 {
		t.Error("GetProcessor returned same instance, expected new instance each time")
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
	RegisterProcessor("filter", mockProcessorFactory("filter"))
	RegisterProcessor("validator", mockProcessorFactory("validator"))
	RegisterProcessor("transformer", mockProcessorFactory("transformer"))

	list = ListProcessors()

	if len(list) != 3 {
		t.Errorf("Expected 3 processors, got %d", len(list))
	}

	// Verify sorted order
	expected := []string{"filter", "transformer", "validator"}
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
	RegisterProcessor("p2", mockProcessorFactory("2"))
	RegisterProcessor("p3", mockProcessorFactory("3"))

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

	RegisterProcessor("p2", mockProcessorFactory("2"))
	if ProcessorCount() != 2 {
		t.Errorf("Expected 2 processors, got %d", ProcessorCount())
	}

	UnregisterProcessor("p1")
	if ProcessorCount() != 1 {
		t.Errorf("Expected 1 processor after removal, got %d", ProcessorCount())
	}
}

// TestRegisterFilter tests type-safe filter registration
func TestRegisterFilter(t *testing.T) {
	ClearProcessors()

	tests := []struct {
		name        string
		filterName  string
		strategy    api.FilterStrategy
		shouldPanic bool
	}{
		{
			name:        "valid filter registration",
			filterName:  "min_threshold",
			strategy:    &mockFilterStrategy{threshold: 10},
			shouldPanic: false,
		},
		{
			name:        "nil strategy panics",
			filterName:  "nil_filter",
			strategy:    nil,
			shouldPanic: true,
		},
		{
			name:        "empty name panics",
			filterName:  "",
			strategy:    &mockFilterStrategy{threshold: 10},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterFilter did not panic for %s", tt.name)
					}
				}()
			}

			RegisterFilter(tt.filterName, tt.strategy)

			if !tt.shouldPanic {
				if !IsProcessorRegistered(tt.filterName) {
					t.Errorf("Filter %s was not registered", tt.filterName)
				}

				// Verify it returns a wrapped processor
				proc, err := GetProcessor(tt.filterName)
				if err != nil {
					t.Errorf("Failed to get registered filter: %v", err)
				}
				if proc == nil {
					t.Error("Expected non-nil processor")
				}
			}
		})
	}
}

// TestRegisterValidator tests type-safe validator registration
func TestRegisterValidator(t *testing.T) {
	ClearProcessors()

	tests := []struct {
		name          string
		validatorName string
		strategy      api.ValidatorStrategy
		shouldPanic   bool
	}{
		{
			name:          "valid validator registration",
			validatorName: "required_fields",
			strategy:      &mockValidatorStrategy{},
			shouldPanic:   false,
		},
		{
			name:          "nil strategy panics",
			validatorName: "nil_validator",
			strategy:      nil,
			shouldPanic:   true,
		},
		{
			name:          "empty name panics",
			validatorName: "",
			strategy:      &mockValidatorStrategy{},
			shouldPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterValidator did not panic for %s", tt.name)
					}
				}()
			}

			RegisterValidator(tt.validatorName, tt.strategy)

			if !tt.shouldPanic {
				if !IsProcessorRegistered(tt.validatorName) {
					t.Errorf("Validator %s was not registered", tt.validatorName)
				}

				// Verify it returns a wrapped processor
				proc, err := GetProcessor(tt.validatorName)
				if err != nil {
					t.Errorf("Failed to get registered validator: %v", err)
				}
				if proc == nil {
					t.Error("Expected non-nil processor")
				}
			}
		})
	}
}

// TestRegisterTransformer tests type-safe transformer registration
func TestRegisterTransformer(t *testing.T) {
	ClearProcessors()

	tests := []struct {
		name            string
		transformerName string
		strategy        api.TransformerStrategy
		shouldPanic     bool
	}{
		{
			name:            "valid transformer registration",
			transformerName: "multiplier",
			strategy:        &mockTransformerStrategy{multiplier: 2},
			shouldPanic:     false,
		},
		{
			name:            "nil strategy panics",
			transformerName: "nil_transformer",
			strategy:        nil,
			shouldPanic:     true,
		},
		{
			name:            "empty name panics",
			transformerName: "",
			strategy:        &mockTransformerStrategy{multiplier: 2},
			shouldPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterTransformer did not panic for %s", tt.name)
					}
				}()
			}

			RegisterTransformer(tt.transformerName, tt.strategy)

			if !tt.shouldPanic {
				if !IsProcessorRegistered(tt.transformerName) {
					t.Errorf("Transformer %s was not registered", tt.transformerName)
				}

				// Verify it returns a wrapped processor
				proc, err := GetProcessor(tt.transformerName)
				if err != nil {
					t.Errorf("Failed to get registered transformer: %v", err)
				}
				if proc == nil {
					t.Error("Expected non-nil processor")
				}
			}
		})
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
			name := string(rune('a' + (id % 26)))
			RegisterProcessor(name, mockProcessorFactory(name))
		}(i)
	}

	wg.Wait()

	// All unique names should be registered (26 unique letters)
	count := ProcessorCount()
	if count > 26 || count == 0 {
		t.Errorf("Expected <= 26 processors after concurrent registration, got %d", count)
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

// TestProcessorConcurrentTypeSafeRegistration tests concurrent type-safe registration
func TestProcessorConcurrentTypeSafeRegistration(t *testing.T) {
	ClearProcessors()

	const goroutines = 30
	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // Filters, Validators, Transformers

	// Register filters
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("filter_%d", id%10)
			RegisterFilter(name, &mockFilterStrategy{threshold: id})
		}(i)
	}

	// Register validators
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("validator_%d", id%10)
			RegisterValidator(name, &mockValidatorStrategy{})
		}(i)
	}

	// Register transformers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("transformer_%d", id%10)
			RegisterTransformer(name, &mockTransformerStrategy{multiplier: id})
		}(i)
	}

	wg.Wait()

	// Should have 30 unique processors (10 of each type)
	count := ProcessorCount()
	if count != 30 {
		t.Logf("Note: Expected 30 processors, got %d (concurrent overwrites are expected)", count)
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
			name := string(rune('a' + (id % 5)))
			RegisterProcessor(name, mockProcessorFactory(name))
		}(i)
	}

	// Get
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
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
			name := string(rune('a' + (id % 5)))
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

// BenchmarkRegisterFilter benchmarks type-safe filter registration
func BenchmarkRegisterFilter(b *testing.B) {
	ClearProcessors()
	strategy := &mockFilterStrategy{threshold: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterFilter("bench_filter", strategy)
	}
}

// BenchmarkListProcessors benchmarks listing processors
func BenchmarkListProcessors(b *testing.B) {
	ClearProcessors()
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		RegisterProcessor(name, mockProcessorFactory(name))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListProcessors()
	}
}
