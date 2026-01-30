package registry

import (
	"context"
	"sync"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// mockProvider is a simple test implementation of ProviderStrategy
type mockProvider struct {
	name string
}

func (m *mockProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	// Simulate data fetching behavior
	return []map[string]interface{}{
		{"id": 1, "name": m.name},
	}, nil
}

// mockProviderFactory creates a new mock provider
func mockProviderFactory(name string) ProviderFactory {
	return func() provider.ProviderStrategy {
		return &mockProvider{name: name}
	}
}

// TestRegisterProvider tests basic registration functionality
func TestRegisterProvider(t *testing.T) {
	// Clean state for test
	ClearProviders()

	tests := []struct {
		name         string
		providerName string
		factory      ProviderFactory
		shouldPanic  bool
	}{
		{
			name:         "valid registration",
			providerName: "test_mock",
			factory:      mockProviderFactory("mock"),
			shouldPanic:  false,
		},
		{
			name:         "empty name panics",
			providerName: "",
			factory:      mockProviderFactory("mock"),
			shouldPanic:  true,
		},
		{
			name:         "nil factory panics",
			providerName: "test_nil",
			factory:      nil,
			shouldPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("RegisterProvider did not panic for %s", tt.name)
					}
				}()
			}

			RegisterProvider(tt.providerName, tt.factory)

			if !tt.shouldPanic {
				if !IsProviderRegistered(tt.providerName) {
					t.Errorf("Provider %s was not registered", tt.providerName)
				}
			}
		})
	}
}

// TestGetProvider tests retrieval of registered providers
func TestGetProvider(t *testing.T) {
	ClearProviders()

	// Register test provider
	RegisterProvider("test_provider", mockProviderFactory("test"))

	tests := []struct {
		name        string
		lookup      string
		shouldError bool
		errorType   string
	}{
		{
			name:        "get existing provider",
			lookup:      "test_provider",
			shouldError: false,
		},
		{
			name:        "get non-existent provider",
			lookup:      "nonexistent",
			shouldError: true,
			errorType:   "*registry.ErrProviderNotFound",
		},
		{
			name:        "empty name returns error",
			lookup:      "",
			shouldError: true,
			errorType:   "ErrEmptyProviderName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetProvider(tt.lookup)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.name, err)
				}
				if result == nil {
					t.Errorf("Expected provider instance, got nil")
				}
			}
		})
	}
}

// TestGetProviderReturnsNewInstance verifies that each call returns a new instance
func TestGetProviderReturnsNewInstance(t *testing.T) {
	ClearProviders()

	RegisterProvider("test", mockProviderFactory("test"))

	p1, err1 := GetProvider("test")
	p2, err2 := GetProvider("test")

	if err1 != nil || err2 != nil {
		t.Fatalf("Unexpected errors: %v, %v", err1, err2)
	}

	// Different instances should have different addresses
	if p1 == p2 {
		t.Error("GetProvider returned same instance, expected new instance each time")
	}
}

// TestListProviders tests listing of registered providers
func TestListProviders(t *testing.T) {
	ClearProviders()

	// Empty list
	list := ListProviders()
	if len(list) != 0 {
		t.Errorf("Expected empty list, got %v", list)
	}

	// Add providers
	RegisterProvider("mock", mockProviderFactory("mock"))
	RegisterProvider("csv", mockProviderFactory("csv"))
	RegisterProvider("postgres", mockProviderFactory("postgres"))

	list = ListProviders()

	if len(list) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(list))
	}

	// Verify sorted order
	expected := []string{"csv", "mock", "postgres"}
	for i, name := range expected {
		if list[i] != name {
			t.Errorf("Expected providers[%d] = %s, got %s", i, name, list[i])
		}
	}
}

// TestIsProviderRegistered tests checking provider existence
func TestIsProviderRegistered(t *testing.T) {
	ClearProviders()

	RegisterProvider("existing", mockProviderFactory("test"))

	tests := []struct {
		name     string
		lookup   string
		expected bool
	}{
		{"existing provider", "existing", true},
		{"non-existent provider", "nonexistent", false},
		{"empty name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsProviderRegistered(tt.lookup)
			if result != tt.expected {
				t.Errorf("IsProviderRegistered(%s) = %v, expected %v", tt.lookup, result, tt.expected)
			}
		})
	}
}

// TestUnregisterProvider tests removal of providers
func TestUnregisterProvider(t *testing.T) {
	ClearProviders()

	RegisterProvider("to_remove", mockProviderFactory("test"))

	if !IsProviderRegistered("to_remove") {
		t.Fatal("Provider should be registered before removal")
	}

	UnregisterProvider("to_remove")

	if IsProviderRegistered("to_remove") {
		t.Error("Provider should not be registered after removal")
	}

	// Unregistering non-existent provider should not panic
	UnregisterProvider("nonexistent")
}

// TestClearProviders tests clearing all providers
func TestClearProviders(t *testing.T) {
	ClearProviders()

	RegisterProvider("p1", mockProviderFactory("1"))
	RegisterProvider("p2", mockProviderFactory("2"))
	RegisterProvider("p3", mockProviderFactory("3"))

	if ProviderCount() != 3 {
		t.Fatalf("Expected 3 providers, got %d", ProviderCount())
	}

	ClearProviders()

	if ProviderCount() != 0 {
		t.Errorf("Expected 0 providers after clear, got %d", ProviderCount())
	}
}

// TestProviderCount tests counting registered providers
func TestProviderCount(t *testing.T) {
	ClearProviders()

	if ProviderCount() != 0 {
		t.Errorf("Expected 0 providers initially, got %d", ProviderCount())
	}

	RegisterProvider("p1", mockProviderFactory("1"))
	if ProviderCount() != 1 {
		t.Errorf("Expected 1 provider, got %d", ProviderCount())
	}

	RegisterProvider("p2", mockProviderFactory("2"))
	if ProviderCount() != 2 {
		t.Errorf("Expected 2 providers, got %d", ProviderCount())
	}

	UnregisterProvider("p1")
	if ProviderCount() != 1 {
		t.Errorf("Expected 1 provider after removal, got %d", ProviderCount())
	}
}

// TestOverwriteProviderRegistration tests that re-registering overwrites previous registration
func TestOverwriteProviderRegistration(t *testing.T) {
	ClearProviders()

	factory1 := mockProviderFactory("first")
	factory2 := mockProviderFactory("second")

	RegisterProvider("test", factory1)
	p1, _ := GetProvider("test")

	RegisterProvider("test", factory2)
	p2, _ := GetProvider("test")

	// Verify different instances by comparing their data
	m1, ok1 := p1.(*mockProvider)
	m2, ok2 := p2.(*mockProvider)

	if !ok1 || !ok2 {
		t.Fatal("Failed to cast to mockProvider")
	}

	if m1.name == m2.name {
		t.Error("Expected different factory output after overwrite")
	}
}

// TestProviderConcurrentRegister tests concurrent registration
func TestProviderConcurrentRegister(t *testing.T) {
	ClearProviders()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 26)))
			RegisterProvider(name, mockProviderFactory(name))
		}(i)
	}

	wg.Wait()

	// All unique names should be registered (26 unique letters)
	count := ProviderCount()
	if count > 26 || count == 0 {
		t.Errorf("Expected <= 26 providers after concurrent registration, got %d", count)
	}
}

// TestProviderConcurrentGet tests concurrent retrieval
func TestProviderConcurrentGet(t *testing.T) {
	ClearProviders()

	RegisterProvider("test", mockProviderFactory("test"))

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := GetProvider("test")
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

// TestProviderConcurrentRegisterAndGet tests concurrent registration and retrieval
func TestProviderConcurrentRegisterAndGet(t *testing.T) {
	ClearProviders()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Half register, half get

	errors := make(chan error, goroutines)

	// Registerers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 10)))
			RegisterProvider(name, mockProviderFactory(name))
		}(i)
	}

	// Getters
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 10)))
			_, err := GetProvider(name)
			// It's OK if not found yet, but shouldn't panic
			if err != nil && err != ErrEmptyProviderName {
				_ = err // Expected - provider might not be registered yet
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

// TestProviderConcurrentMixedOperations tests all operations concurrently
func TestProviderConcurrentMixedOperations(t *testing.T) {
	ClearProviders()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 5) // 5 different operations

	// Register
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			RegisterProvider(name, mockProviderFactory(name))
		}(i)
	}

	// Get
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			_, _ = GetProvider(name)
		}(i)
	}

	// List
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ListProviders()
		}()
	}

	// IsRegistered
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := string(rune('a' + (id % 5)))
			IsProviderRegistered(name)
		}(i)
	}

	// Count
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ProviderCount()
		}()
	}

	wg.Wait()

	// If we get here without deadlock or panic, test passes
	t.Logf("Successfully completed %d concurrent operations", goroutines*5)
}

// TestProviderRaceDetector should be run with -race flag
func TestProviderRaceDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detector test in short mode")
	}

	ClearProviders()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			RegisterProvider("test", mockProviderFactory("test"))
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			_, _ = GetProvider("test")
			ListProviders()
			IsProviderRegistered("test")
			ProviderCount()
		}
		done <- true
	}()

	<-done
	<-done
}

// BenchmarkRegisterProvider benchmarks provider registration
func BenchmarkRegisterProvider(b *testing.B) {
	ClearProviders()
	factory := mockProviderFactory("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RegisterProvider("bench_provider", factory)
	}
}

// BenchmarkGetProvider benchmarks provider retrieval
func BenchmarkGetProvider(b *testing.B) {
	ClearProviders()
	RegisterProvider("bench", mockProviderFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetProvider("bench")
	}
}

// BenchmarkGetProviderParallel benchmarks concurrent provider retrieval
func BenchmarkGetProviderParallel(b *testing.B) {
	ClearProviders()
	RegisterProvider("bench", mockProviderFactory("test"))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = GetProvider("bench")
		}
	})
}

// BenchmarkListProviders benchmarks listing providers
func BenchmarkListProviders(b *testing.B) {
	ClearProviders()
	for i := 0; i < 10; i++ {
		name := string(rune('a' + i))
		RegisterProvider(name, mockProviderFactory(name))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListProviders()
	}
}

// BenchmarkIsProviderRegistered benchmarks existence checking
func BenchmarkIsProviderRegistered(b *testing.B) {
	ClearProviders()
	RegisterProvider("bench", mockProviderFactory("test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsProviderRegistered("bench")
	}
}
