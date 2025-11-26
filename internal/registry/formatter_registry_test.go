package registry

import (
	"context"
	"sync"
	"testing"

	"github.com/AshishBagdane/report-engine/internal/formatter"
)

// mockFormatter is a simple test implementation of FormatStrategy
type mockFormatter struct {
	name string
}

func (m *mockFormatter) Format(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
	return []byte(m.name), nil
}

// mockFormatterFactory creates a new mock formatter
func mockFormatterFactory(name string) FormatterFactory {
	return func() formatter.FormatStrategy {
		return &mockFormatter{name: name}
	}
}

// TestRegisterFormatter tests basic registration functionality
func TestRegisterFormatter(t *testing.T) {
	// Clean state for test
	ClearFormatters()

	tests := []struct {
		name          string
		formatterName string
		factory       FormatterFactory
		shouldPanic   bool
	}{
		{
			name:          "valid registration",
			formatterName: "test_json",
			factory:       mockFormatterFactory("json"),
			shouldPanic:   false,
		},
		{
			name:          "empty name panics",
			formatterName: "",
			factory:       mockFormatterFactory("json"),
			shouldPanic:   true,
		},
		{
			name:          "nil factory panics",
			formatterName: "test_nil",
			factory:       nil,
			shouldPanic:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterFormatter did not panic for %s", tt.name)
					}
				}()
			}

			RegisterFormatter(tt.formatterName, tt.factory)

			if !tt.shouldPanic {
				if !IsFormatterRegistered(tt.formatterName) {
					t.Errorf("Formatter %s was not registered", tt.formatterName)
				}
			}
		})
	}
}

// TestGetFormatter tests retrieval of registered formatters
func TestGetFormatter(t *testing.T) {
	ClearFormatters()

	// Register test formatter
	RegisterFormatter("test_formatter", mockFormatterFactory("test"))

	tests := []struct {
		name        string
		lookup      string
		shouldError bool
		errorType   string
	}{
		{
			name:        "get existing formatter",
			lookup:      "test_formatter",
			shouldError: false,
		},
		{
			name:        "get non-existent formatter",
			lookup:      "nonexistent",
			shouldError: true,
			errorType:   "*registry.ErrFormatterNotFound",
		},
		{
			name:        "empty name returns error",
			lookup:      "",
			shouldError: true,
			errorType:   "ErrEmptyFormatterName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetFormatter(tt.lookup)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
				if result == nil {
					t.Errorf("Expected formatter instance, got nil")
				}
			}
		})
	}
}

// TestGetFormatterReturnsNewInstance verifies that each call returns a new instance
func TestGetFormatterReturnsNewInstance(t *testing.T) {
	ClearFormatters()

	RegisterFormatter("test", mockFormatterFactory("test"))

	f1, err1 := GetFormatter("test")
	f2, err2 := GetFormatter("test")

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected errors: %v, %v", err1, err2)
	}

	// Different instances should have different addresses
	if f1 == f2 {
		t.Error("GetFormatter returned same instance, expected new instance each time")
	}
}

// TestListFormatters tests listing of registered formatters
func TestListFormatters(t *testing.T) {
	ClearFormatters()

	// Empty list
	list := ListFormatters()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %v", list)
	}

	// Add formatters
	RegisterFormatter("json", mockFormatterFactory("json"))
	RegisterFormatter("csv", mockFormatterFactory("csv"))
	RegisterFormatter("xml", mockFormatterFactory("xml"))

	list = ListFormatters()

	if len(list) != 3 {
		t.Errorf("Expected 3 formatters, got %d", len(list))
	}

	// Verify sorted order
	expected := []string{"csv", "json", "xml"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("Expected formatters[%d] = %s, got %s", i, name, list[i])
		}
	}
}

// TestIsFormatterRegistered tests checking formatter existence
func TestIsFormatterRegistered(t *testing.T) {
	ClearFormatters()

	RegisterFormatter("existing", mockFormatterFactory("test"))

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"existing formatter", "existing", true},
		{"non-existent formatter", "nonexistent", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFormatterRegistered(tt.lookup)
			if result != tt.expected {
				t.Errorf("IsFormatterRegistered(%s) = %v, expected %v", tt.lookup, result, tt.expected)
			}
		})
	}
}

// TestUnregisterFormatter tests removal of formatters
func TestUnregisterFormatter(t *testing.T) {
	ClearFormatters()

	RegisterFormatter("to_remove", mockFormatterFactory("test"))

	if !IsFormatterRegistered("to_remove") {
		t.Fatal("Formatter should be registered before removal")
	}

	UnregisterFormatter("to_remove")

	if IsFormatterRegistered("to_remove") {
		t.Error("Formatter should not be registered after removal")
	}

	// Unregistering non-existent formatter should not panic
	UnregisterFormatter("nonexistent")
}

// TestClearFormatters tests clearing all formatters
func TestClearFormatters(t *testing.T) {
	ClearFormatters()

	RegisterFormatter("f1", mockFormatterFactory("1"))
	RegisterFormatter("f2", mockFormatterFactory("2"))
	RegisterFormatter("f3", mockFormatterFactory("3"))

	if FormatterCount() != 3 {
		t.Fatalf("Expected 3 formatters, got %d", FormatterCount())
	}

	ClearFormatters()

	if FormatterCount() != 0 {
		t.Errorf("Expected 0 formatters after clear, got %d", FormatterCount())
	}
}

// TestFormatterCount tests counting registered formatters
func TestFormatterCount(t *testing.T) {
	ClearFormatters()

	if FormatterCount() != 0 {
		t.Errorf("Expected 0 formatters initially, got %d", FormatterCount())
	}

	RegisterFormatter("f1", mockFormatterFactory("1"))
	if FormatterCount() != 1 {
		t.Errorf("Expected 1 formatter, got %d", FormatterCount())
	}

	RegisterFormatter("f2", mockFormatterFactory("2"))
	if FormatterCount() != 2 {
		t.Errorf("Expected 2 formatters, got %d", FormatterCount())
	}

	UnregisterFormatter("f1")
	if FormatterCount() != 1 {
		t.Errorf("Expected 1 formatter after removal, got %d", FormatterCount())
	}
}

// TestOverwriteRegistration tests that re-registering overwrites previous registration
func TestOverwriteRegistration(t *testing.T) {
	ClearFormatters()

	factory1 := mockFormatterFactory("first")
	factory2 := mockFormatterFactory("second")

	RegisterFormatter("test", factory1)
	f1, _ := GetFormatter("test")

	RegisterFormatter("test", factory2)
	f2, _ := GetFormatter("test")

	// Get actual formatted output to verify different factories
	ctx := context.Background()
	out1, _ := f1.Format(ctx, nil)
	out2, _ := f2.Format(ctx, nil)

	if string(out1) == string(out2) {
		t.Error("Expected different factory output after overwrite")
	}
}

// TestFormatterConcurrentRegister tests concurrent registration
func TestFormatterConcurrentRegister(t *testing.T) {
	ClearFormatters()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 26)))
			RegisterFormatter(name, mockFormatterFactory(name))
		}(i)
	}

	wg.Wait()

	// All unique names should be registered (26 unique letters)
	count := FormatterCount()
	if count > 26 || count == 0 {
		t.Errorf("Expected <= 26 formatters after concurrent registration, got %d", count)
	}
}

// TestFormatterConcurrentGet tests concurrent retrieval
func TestFormatterConcurrentGet(t *testing.T) {
	ClearFormatters()

	RegisterFormatter("test", mockFormatterFactory("test"))

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := GetFormatter("test")
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

// TestFormatterConcurrentRegisterAndGet tests concurrent registration and retrieval
func TestFormatterConcurrentRegisterAndGet(t *testing.T) {
	ClearFormatters()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Half register, half get

	errors := make(chan error, goroutines)

	// Registerers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 10)))
			RegisterFormatter(name, mockFormatterFactory(name))
		}(i)
	}

	// Getters
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 10)))
			_, err := GetFormatter(name)
			// It's OK if not found yet, but shouldn't panic
			if err != nil && err != ErrEmptyFormatterName {
				// Expected - formatter might not be registered yet
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

// TestFormatterConcurrentMixedOperations tests all operations concurrently
func TestFormatterConcurrentMixedOperations(t *testing.T) {
	ClearFormatters()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 5) // 5 different operations

	// Register
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			RegisterFormatter(name, mockFormatterFactory(name))
		}(i)
	}

	// Get
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			GetFormatter(name)
		}(i)
	}

	// List
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ListFormatters()
		}()
	}

	// IsRegistered
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			IsFormatterRegistered(name)
		}(i)
	}

	// Count
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			FormatterCount()
		}()
	}

	wg.Wait()

	// If we get here without deadlock or panic, test passes
	t.Logf("Successfully completed %d concurrent operations", goroutines*5)
}

// TestFormatterRaceDetector should be run with -race flag
func TestFormatterRaceDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detector test in short mode")
	}

	ClearFormatters()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			RegisterFormatter("test", mockFormatterFactory("test"))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			GetFormatter("test")
			ListFormatters()
			IsFormatterRegistered("test")
			FormatterCount()
		}
		done <- true
	}()

	<-done
	<-done
}

// BenchmarkRegisterFormatter benchmarks formatter registration
func BenchmarkRegisterFormatter(b *testing.B) {
	ClearFormatters()
	factory := mockFormatterFactory("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterFormatter("bench_formatter", factory)
	}
}

// BenchmarkGetFormatter benchmarks formatter retrieval
func BenchmarkGetFormatter(b *testing.B) {
	ClearFormatters()
	RegisterFormatter("bench", mockFormatterFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetFormatter("bench")
	}
}

// BenchmarkGetFormatterParallel benchmarks concurrent formatter retrieval
func BenchmarkGetFormatterParallel(b *testing.B) {
	ClearFormatters()
	RegisterFormatter("bench", mockFormatterFactory("test"))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetFormatter("bench")
		}
	})
}

// BenchmarkListFormatters benchmarks listing formatters
func BenchmarkListFormatters(b *testing.B) {
	ClearFormatters()
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		RegisterFormatter(name, mockFormatterFactory(name))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListFormatters()
	}
}

// BenchmarkIsFormatterRegistered benchmarks existence checking
func BenchmarkIsFormatterRegistered(b *testing.B) {
	ClearFormatters()
	RegisterFormatter("bench", mockFormatterFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsFormatterRegistered("bench")
	}
}
