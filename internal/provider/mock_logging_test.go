package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/AshishBagdane/go-report-engine/internal/logging"
)

// TestMockProviderWithLogger tests logger injection
func TestMockProviderWithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelInfo,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock.test",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "score": 95},
		{"id": 2, "name": "Bob", "score": 88},
	}

	provider := NewMockProvider(testData)
	result := provider.WithLogger(logger)

	// Should return provider for chaining
	if result != provider {
		t.Error("WithLogger should return provider for chaining")
	}

	// Logger should be set
	if provider.logger != logger {
		t.Error("Logger was not set correctly")
	}
}

// TestMockProviderGetLoggerDefault tests lazy logger initialization
func TestMockProviderGetLoggerDefault(t *testing.T) {
	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := &MockProvider{
		Data:   testData,
		logger: nil,
	}

	logger := provider.getLogger()

	if logger == nil {
		t.Fatal("getLogger() returned nil")
	}

	// Should create default logger
	if provider.logger == nil {
		t.Error("Logger was not initialized")
	}

	// Calling again should return same logger
	logger2 := provider.getLogger()
	if logger != logger2 {
		t.Error("getLogger() should return same logger instance")
	}
}

// TestMockProviderFetchLogging tests that fetch operations are logged
func TestMockProviderFetchLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelInfo,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	ctx := context.Background()
	_, err := provider.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}

	logs := buf.String()

	// Verify fetch starting log
	if !strings.Contains(logs, "fetch starting") {
		t.Error("Missing 'fetch starting' log entry")
	}

	// Verify fetch completed log
	if !strings.Contains(logs, "fetch completed") {
		t.Error("Missing 'fetch completed' log entry")
	}

	// Verify provider_type is logged
	if !strings.Contains(logs, "provider_type") {
		t.Error("Missing 'provider_type' field in logs")
	}

	// Verify duration metrics are logged
	if !strings.Contains(logs, "duration_ms") {
		t.Error("Missing 'duration_ms' field in logs")
	}

	// Verify record count is logged
	if !strings.Contains(logs, "record_count") {
		t.Error("Missing 'record_count' field in logs")
	}
}

// TestMockProviderFetchLoggingDebug tests debug level logging
func TestMockProviderFetchLoggingDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelDebug,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	ctx := context.Background()
	_, err := provider.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}

	logs := buf.String()

	// At debug level, should have additional details
	if !strings.Contains(logs, "fetch data structure") {
		t.Error("Missing debug-level 'fetch data structure' log entry")
	}

	// Should log field names
	if !strings.Contains(logs, "fields") {
		t.Error("Missing 'fields' in debug logs")
	}
}

// TestMockProviderFetchLoggingMetrics tests logged metrics are accurate
func TestMockProviderFetchLoggingMetrics(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelInfo,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	ctx := context.Background()
	data, err := provider.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}

	// Parse logs to verify metrics
	logs := strings.Split(buf.String(), "\n")

	var foundCompletionLog bool
	for _, line := range logs {
		if strings.Contains(line, "fetch completed") {
			foundCompletionLog = true

			// Parse JSON log entry
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
				continue
			}

			// Verify record count matches actual data
			if recordCount, ok := logEntry["record_count"].(float64); ok {
				if int(recordCount) != len(data) {
					t.Errorf("Logged record_count = %d, expected %d", int(recordCount), len(data))
				}
			} else {
				t.Error("record_count not found or wrong type in log")
			}

			// Verify duration_ms exists and is reasonable
			if durationMs, ok := logEntry["duration_ms"].(float64); ok {
				if durationMs < 0 {
					t.Error("duration_ms should not be negative")
				}
				// Mock provider should be very fast
				if durationMs > 1000 {
					t.Errorf("duration_ms = %f, seems too high for mock provider", durationMs)
				}
			} else {
				t.Error("duration_ms not found or wrong type in log")
			}

			break
		}
	}

	if !foundCompletionLog {
		t.Error("Did not find fetch completion log entry")
	}
}

// TestMockProviderFetchLoggingTextFormat tests text format logging
func TestMockProviderFetchLoggingTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelInfo,
		Format:    logging.FormatText,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	ctx := context.Background()
	_, err := provider.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}

	logs := buf.String()

	// Text format should still contain key information
	if !strings.Contains(logs, "fetch starting") {
		t.Error("Missing 'fetch starting' in text logs")
	}

	if !strings.Contains(logs, "fetch completed") {
		t.Error("Missing 'fetch completed' in text logs")
	}

	if !strings.Contains(logs, "record_count") {
		t.Error("Missing 'record_count' in text logs")
	}
}

// TestMockProviderConcurrentFetchLogging tests logging under concurrent access
func TestMockProviderConcurrentFetchLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelInfo,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	const goroutines = 10
	errors := make(chan error, goroutines)
	done := make(chan struct{}, goroutines)

	ctx := context.Background()

	// Launch concurrent fetches
	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := provider.Fetch(ctx)
			if err != nil {
				errors <- err
			}
			done <- struct{}{}
		}()
	}

	// Wait for completion
	for i := 0; i < goroutines; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent Fetch() failed: %v", err)
		case <-done:
			// Success
		}
	}

	logs := buf.String()

	// Should have logs from all goroutines
	startingCount := strings.Count(logs, "fetch starting")
	completedCount := strings.Count(logs, "fetch completed")

	if startingCount != goroutines {
		t.Errorf("Found %d 'fetch starting' logs, expected %d", startingCount, goroutines)
	}

	if completedCount != goroutines {
		t.Errorf("Found %d 'fetch completed' logs, expected %d", completedCount, goroutines)
	}
}

// TestMockProviderLoggingWithoutExplicitLogger tests default logger behavior
func TestMockProviderLoggingWithoutExplicitLogger(t *testing.T) {
	// Create provider without setting logger
	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData)

	ctx := context.Background()
	// Should not panic and should create default logger
	data, err := provider.Fetch(ctx)

	if err != nil {
		t.Fatalf("Fetch() returned error: %v", err)
	}

	if data == nil {
		t.Fatal("Fetch() returned nil data")
	}

	// Logger should have been initialized
	if provider.logger == nil {
		t.Error("Logger should have been lazily initialized")
	}
}

// BenchmarkMockProviderFetchWithLogging benchmarks fetch with logging enabled
func BenchmarkMockProviderFetchWithLogging(b *testing.B) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelInfo,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Fetch(ctx)
	}
}

// BenchmarkMockProviderFetchWithoutLogging benchmarks fetch without explicit logger
func BenchmarkMockProviderFetchWithoutLogging(b *testing.B) {
	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Fetch(ctx)
	}
}

// BenchmarkMockProviderFetchDebugLogging benchmarks with debug level logging
func BenchmarkMockProviderFetchDebugLogging(b *testing.B) {
	var buf bytes.Buffer
	logger := logging.NewLogger(logging.Config{
		Level:     logging.LevelDebug,
		Format:    logging.FormatJSON,
		Output:    &buf,
		Component: "provider.mock",
	})

	testData := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
	}

	provider := NewMockProvider(testData).WithLogger(logger)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.Fetch(ctx)
	}
}
