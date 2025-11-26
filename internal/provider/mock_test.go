package provider

import (
	"context"
	"testing"
)

// Default test data for tests
var testData = []map[string]interface{}{
	{"id": 1, "name": "Alice", "score": 95},
	{"id": 2, "name": "Bob", "score": 88},
}

// TestNewMockProvider tests the factory function
func TestNewMockProvider(t *testing.T) {
	provider := NewMockProvider(testData)

	if provider == nil {
		t.Fatal("NewMockProvider() returned nil")
	}

	// Verify it implements ProviderStrategy
	var _ ProviderStrategy = provider
}

// TestMockProviderFetch tests basic fetch functionality
func TestMockProviderFetch(t *testing.T) {
	provider := NewMockProvider(testData)

	ctx := context.Background()
	data, err := provider.Fetch(ctx)

	if err != nil {
		t.Fatalf("Fetch() returned unexpected error: %v", err)
	}

	if data == nil {
		t.Fatal("Fetch() returned nil data")
	}

	// Verify we get the expected mock data
	expectedRecords := 2
	if len(data) != expectedRecords {
		t.Errorf("Fetch() returned %d records, expected %d", len(data), expectedRecords)
	}
}

// TestMockProviderDataStructure tests the structure of returned data
func TestMockProviderDataStructure(t *testing.T) {
	provider := NewMockProvider(testData)

	ctx := context.Background()
	data, err := provider.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	// Test first record
	if len(data) < 1 {
		t.Fatal("Expected at least 1 record")
	}

	record := data[0]

	// Check for expected fields
	expectedFields := []string{"id", "name", "score"}
	for _, field := range expectedFields {
		if _, ok := record[field]; !ok {
			t.Errorf("Record missing expected field: %s", field)
		}
	}

	// Verify field types
	if _, ok := record["id"].(int); !ok {
		t.Error("Field 'id' should be int")
	}

	if _, ok := record["name"].(string); !ok {
		t.Error("Field 'name' should be string")
	}

	if _, ok := record["score"].(int); !ok {
		t.Error("Field 'score' should be int")
	}
}

// TestMockProviderDataValues tests specific data values
func TestMockProviderDataValues(t *testing.T) {
	provider := NewMockProvider(testData)

	ctx := context.Background()
	data, err := provider.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	tests := []struct {
		name          string
		recordIndex   int
		expectedID    int
		expectedName  string
		expectedScore int
	}{
		{
			name:          "first record",
			recordIndex:   0,
			expectedID:    1,
			expectedName:  "Alice",
			expectedScore: 95,
		},
		{
			name:          "second record",
			recordIndex:   1,
			expectedID:    2,
			expectedName:  "Bob",
			expectedScore: 88,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.recordIndex >= len(data) {
				t.Fatalf("Not enough records, need at least %d", tt.recordIndex+1)
			}

			record := data[tt.recordIndex]

			if id := record["id"].(int); id != tt.expectedID {
				t.Errorf("Record %d: id = %d, expected %d", tt.recordIndex, id, tt.expectedID)
			}

			if name := record["name"].(string); name != tt.expectedName {
				t.Errorf("Record %d: name = %s, expected %s", tt.recordIndex, name, tt.expectedName)
			}

			if score := record["score"].(int); score != tt.expectedScore {
				t.Errorf("Record %d: score = %d, expected %d", tt.recordIndex, score, tt.expectedScore)
			}
		})
	}
}

// TestMockProviderMultipleCalls tests that multiple calls work correctly
func TestMockProviderMultipleCalls(t *testing.T) {
	provider := NewMockProvider(testData)

	ctx := context.Background()

	// First call
	data1, err1 := provider.Fetch(ctx)
	if err1 != nil {
		t.Fatalf("First Fetch() failed: %v", err1)
	}

	// Second call
	data2, err2 := provider.Fetch(ctx)
	if err2 != nil {
		t.Fatalf("Second Fetch() failed: %v", err2)
	}

	// Both calls should return data
	if len(data1) != len(data2) {
		t.Errorf("Multiple Fetch() calls returned different lengths: %d vs %d", len(data1), len(data2))
	}

	// Verify data consistency (same values each time)
	if len(data1) > 0 && len(data2) > 0 {
		if data1[0]["id"] != data2[0]["id"] {
			t.Error("Multiple Fetch() calls should return consistent data")
		}
	}
}

// TestMockProviderContextCancellation tests context cancellation
func TestMockProviderContextCancellation(t *testing.T) {
	provider := NewMockProvider(testData)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.Fetch(ctx)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// TestMockProviderConcurrentAccess tests concurrent calls to Fetch
func TestMockProviderConcurrentAccess(t *testing.T) {
	provider := NewMockProvider(testData)

	const goroutines = 10
	errors := make(chan error, goroutines)
	results := make(chan []map[string]interface{}, goroutines)

	ctx := context.Background()

	// Launch concurrent fetches
	for i := 0; i < goroutines; i++ {
		go func() {
			data, err := provider.Fetch(ctx)
			if err != nil {
				errors <- err
				return
			}
			results <- data
		}()
	}

	// Collect results
	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent Fetch() failed: %v", err)
		case data := <-results:
			if data == nil {
				t.Error("Concurrent Fetch() returned nil")
			}
			if len(data) != 2 {
				t.Errorf("Concurrent Fetch() returned %d records, expected 2", len(data))
			}
		}
	}
}

// TestMockProviderImplementsInterface verifies interface implementation
func TestMockProviderImplementsInterface(t *testing.T) {
	var _ ProviderStrategy = (*MockProvider)(nil)
}

// TestMockProviderZeroValue tests behavior with zero value
func TestMockProviderZeroValue(t *testing.T) {
	// Create zero value (not via factory)
	var provider MockProvider

	ctx := context.Background()
	data, err := provider.Fetch(ctx)

	if err != nil {
		t.Errorf("Zero value Fetch() returned error: %v", err)
	}

	// Zero value has nil Data, should return nil
	if data != nil {
		t.Error("Zero value Fetch() should return nil")
	}
}

// TestMockProviderEmptyData tests with empty data
func TestMockProviderEmptyData(t *testing.T) {
	provider := NewMockProvider([]map[string]interface{}{})

	ctx := context.Background()
	data, err := provider.Fetch(ctx)

	if err != nil {
		t.Errorf("Fetch() with empty data returned error: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Fetch() returned %d records, expected 0", len(data))
	}
}

// BenchmarkMockProviderFetch benchmarks the Fetch operation
func BenchmarkMockProviderFetch(b *testing.B) {
	provider := NewMockProvider(testData)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Fetch(ctx)
	}
}

// BenchmarkMockProviderFetchParallel benchmarks concurrent Fetch operations
func BenchmarkMockProviderFetchParallel(b *testing.B) {
	provider := NewMockProvider(testData)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			provider.Fetch(ctx)
		}
	})
}

// BenchmarkNewMockProvider benchmarks provider creation
func BenchmarkNewMockProvider(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMockProvider(testData)
	}
}
