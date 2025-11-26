package registry

import (
	"context"
	"sync"
	"testing"

	"github.com/AshishBagdane/report-engine/internal/output"
)

// mockOutput is a simple test implementation of OutputStrategy
type mockOutput struct {
	name string
}

func (m *mockOutput) Send(ctx context.Context, data []byte) error {
	// Simulate output behavior
	return nil
}

// mockOutputFactory creates a new mock output
func mockOutputFactory(name string) OutputFactory {
	return func() output.OutputStrategy {
		return &mockOutput{name: name}
	}
}

// TestRegisterOutput tests basic registration functionality
func TestRegisterOutput(t *testing.T) {
	// Clean state for test
	ClearOutputs()

	tests := []struct {
		name        string
		outputName  string
		factory     OutputFactory
		shouldPanic bool
	}{
		{
			name:        "valid registration",
			outputName:  "test_console",
			factory:     mockOutputFactory("console"),
			shouldPanic: false,
		},
		{
			name:        "empty name panics",
			outputName:  "",
			factory:     mockOutputFactory("console"),
			shouldPanic: true,
		},
		{
			name:        "nil factory panics",
			outputName:  "test_nil",
			factory:     nil,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterOutput did not panic for %s", tt.name)
					}
				}()
			}

			RegisterOutput(tt.outputName, tt.factory)

			if !tt.shouldPanic {
				if !IsOutputRegistered(tt.outputName) {
					t.Errorf("Output %s was not registered", tt.outputName)
				}
			}
		})
	}
}

// TestGetOutput tests retrieval of registered outputs
func TestGetOutput(t *testing.T) {
	ClearOutputs()

	// Register test output
	RegisterOutput("test_output", mockOutputFactory("test"))

	tests := []struct {
		name        string
		lookup      string
		shouldError bool
		errorType   string
	}{
		{
			name:        "get existing output",
			lookup:      "test_output",
			shouldError: false,
		},
		{
			name:        "get non-existent output",
			lookup:      "nonexistent",
			shouldError: true,
			errorType:   "*registry.ErrOutputNotFound",
		},
		{
			name:        "empty name returns error",
			lookup:      "",
			shouldError: true,
			errorType:   "ErrEmptyOutputName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetOutput(tt.lookup)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
				if result == nil {
					t.Errorf("Expected output instance, got nil")
				}
			}
		})
	}
}

// TestGetOutputReturnsNewInstance verifies that each call returns a new instance
func TestGetOutputReturnsNewInstance(t *testing.T) {
	ClearOutputs()

	RegisterOutput("test", mockOutputFactory("test"))

	o1, err1 := GetOutput("test")
	o2, err2 := GetOutput("test")

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected errors: %v, %v", err1, err2)
	}

	// Different instances should have different addresses
	if o1 == o2 {
		t.Error("GetOutput returned same instance, expected new instance each time")
	}
}

// TestListOutputs tests listing of registered outputs
func TestListOutputs(t *testing.T) {
	ClearOutputs()

	// Empty list
	list := ListOutputs()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %v", list)
	}

	// Add outputs
	RegisterOutput("console", mockOutputFactory("console"))
	RegisterOutput("file", mockOutputFactory("file"))
	RegisterOutput("s3", mockOutputFactory("s3"))

	list = ListOutputs()

	if len(list) != 3 {
		t.Errorf("Expected 3 outputs, got %d", len(list))
	}

	// Verify sorted order
	expected := []string{"console", "file", "s3"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("Expected outputs[%d] = %s, got %s", i, name, list[i])
		}
	}
}

// TestIsOutputRegistered tests checking output existence
func TestIsOutputRegistered(t *testing.T) {
	ClearOutputs()

	RegisterOutput("existing", mockOutputFactory("test"))

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"existing output", "existing", true},
		{"non-existent output", "nonexistent", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOutputRegistered(tt.lookup)
			if result != tt.expected {
				t.Errorf("IsOutputRegistered(%s) = %v, expected %v", tt.lookup, result, tt.expected)
			}
		})
	}
}

// TestUnregisterOutput tests removal of outputs
func TestUnregisterOutput(t *testing.T) {
	ClearOutputs()

	RegisterOutput("to_remove", mockOutputFactory("test"))

	if !IsOutputRegistered("to_remove") {
		t.Fatal("Output should be registered before removal")
	}

	UnregisterOutput("to_remove")

	if IsOutputRegistered("to_remove") {
		t.Error("Output should not be registered after removal")
	}

	// Unregistering non-existent output should not panic
	UnregisterOutput("nonexistent")
}

// TestClearOutputs tests clearing all outputs
func TestClearOutputs(t *testing.T) {
	ClearOutputs()

	RegisterOutput("o1", mockOutputFactory("1"))
	RegisterOutput("o2", mockOutputFactory("2"))
	RegisterOutput("o3", mockOutputFactory("3"))

	if OutputCount() != 3 {
		t.Fatalf("Expected 3 outputs, got %d", OutputCount())
	}

	ClearOutputs()

	if OutputCount() != 0 {
		t.Errorf("Expected 0 outputs after clear, got %d", OutputCount())
	}
}

// TestOutputCount tests counting registered outputs
func TestOutputCount(t *testing.T) {
	ClearOutputs()

	if OutputCount() != 0 {
		t.Errorf("Expected 0 outputs initially, got %d", OutputCount())
	}

	RegisterOutput("o1", mockOutputFactory("1"))
	if OutputCount() != 1 {
		t.Errorf("Expected 1 output, got %d", OutputCount())
	}

	RegisterOutput("o2", mockOutputFactory("2"))
	if OutputCount() != 2 {
		t.Errorf("Expected 2 outputs, got %d", OutputCount())
	}

	UnregisterOutput("o1")
	if OutputCount() != 1 {
		t.Errorf("Expected 1 output after removal, got %d", OutputCount())
	}
}

// TestOverwriteOutputRegistration tests that re-registering overwrites previous registration
func TestOverwriteOutputRegistration(t *testing.T) {
	ClearOutputs()

	factory1 := mockOutputFactory("first")
	factory2 := mockOutputFactory("second")

	RegisterOutput("test", factory1)
	o1, _ := GetOutput("test")

	RegisterOutput("test", factory2)
	o2, _ := GetOutput("test")

	// Verify different instances by comparing their type assertion
	m1, ok1 := o1.(*mockOutput)
	m2, ok2 := o2.(*mockOutput)

	if !ok1 || !ok2 {
		t.Fatal("Failed to cast to mockOutput")
	}

	if m1.name == m2.name {
		t.Error("Expected different factory output after overwrite")
	}
}

// TestOutputConcurrentRegister tests concurrent registration
func TestOutputConcurrentRegister(t *testing.T) {
	ClearOutputs()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 26)))
			RegisterOutput(name, mockOutputFactory(name))
		}(i)
	}

	wg.Wait()

	// All unique names should be registered (26 unique letters)
	count := OutputCount()
	if count > 26 || count == 0 {
		t.Errorf("Expected <= 26 outputs after concurrent registration, got %d", count)
	}
}

// TestOutputConcurrentGet tests concurrent retrieval
func TestOutputConcurrentGet(t *testing.T) {
	ClearOutputs()

	RegisterOutput("test", mockOutputFactory("test"))

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := GetOutput("test")
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

// TestOutputConcurrentRegisterAndGet tests concurrent registration and retrieval
func TestOutputConcurrentRegisterAndGet(t *testing.T) {
	ClearOutputs()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Half register, half get

	errors := make(chan error, goroutines)

	// Registerers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 10)))
			RegisterOutput(name, mockOutputFactory(name))
		}(i)
	}

	// Getters
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 10)))
			_, err := GetOutput(name)
			// It's OK if not found yet, but shouldn't panic
			if err != nil && err != ErrEmptyOutputName {
				// Expected - output might not be registered yet
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// No panics means success
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	}
}

// TestOutputConcurrentMixedOperations tests all operations concurrently
func TestOutputConcurrentMixedOperations(t *testing.T) {
	ClearOutputs()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 5) // 5 different operations

	// Register
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			RegisterOutput(name, mockOutputFactory(name))
		}(i)
	}

	// Get
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			GetOutput(name)
		}(i)
	}

	// List
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ListOutputs()
		}()
	}

	// IsRegistered
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			IsOutputRegistered(name)
		}(i)
	}

	// Count
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			OutputCount()
		}()
	}

	wg.Wait()

	// If we get here without deadlock or panic, test passes
	t.Logf("Successfully completed %d concurrent operations", goroutines*5)
}

// TestOutputRaceDetector should be run with -race flag
func TestOutputRaceDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detector test in short mode")
	}

	ClearOutputs()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			RegisterOutput("test", mockOutputFactory("test"))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			GetOutput("test")
			ListOutputs()
			IsOutputRegistered("test")
			OutputCount()
		}
		done <- true
	}()

	<-done
	<-done
}

// BenchmarkRegisterOutput benchmarks output registration
func BenchmarkRegisterOutput(b *testing.B) {
	ClearOutputs()
	factory := mockOutputFactory("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterOutput("bench_output", factory)
	}
}

// BenchmarkGetOutput benchmarks output retrieval
func BenchmarkGetOutput(b *testing.B) {
	ClearOutputs()
	RegisterOutput("bench", mockOutputFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetOutput("bench")
	}
}

// BenchmarkGetOutputParallel benchmarks concurrent output retrieval
func BenchmarkGetOutputParallel(b *testing.B) {
	ClearOutputs()
	RegisterOutput("bench", mockOutputFactory("test"))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetOutput("bench")
		}
	})
}

// BenchmarkListOutputs benchmarks listing outputs
func BenchmarkListOutputs(b *testing.B) {
	ClearOutputs()
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		RegisterOutput(name, mockOutputFactory(name))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListOutputs()
	}
}

// BenchmarkIsOutputRegistered benchmarks existence checking
func BenchmarkIsOutputRegistered(b *testing.B) {
	ClearOutputs()
	RegisterOutput("bench", mockOutputFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsOutputRegistered("bench")
	}
}
