// Package provider contains data source implementations that fetch raw data
// for the report engine pipeline.
package provider

import (
	"context"
	"time"

	"github.com/AshishBagdane/report-engine/internal/logging"
)

// MockProvider is a test implementation that returns hardcoded sample data.
// It's primarily used for testing, examples, and development. This provider
// includes comprehensive structured logging for observability.
//
// The MockProvider returns two sample records with id, name, and score fields.
// It demonstrates the Provider interface contract and logging patterns.
//
// Thread-safe: Yes. Multiple goroutines can safely call Fetch() concurrently.
//
// Example:
//
//	provider := provider.NewMockProvider()
//
//	// Optional: Set custom logger
//	logger := logging.NewLogger(logging.Config{
//	    Level: logging.LevelDebug,
//	    Component: "provider.mock",
//	})
//	provider.WithLogger(logger)
//
//	data, err := provider.Fetch()
//	if err != nil {
//	    log.Fatal(err)
//	}
type MockProvider struct {
	// logger provides structured logging for fetch operations.
	// If nil, a default logger is created on first use.
	logger *logging.Logger
}

// NewMockProvider creates a new MockProvider with default configuration.
// The returned provider has a default logger that can be overridden using WithLogger().
//
// Returns a ProviderStrategy implementation ready for use in the report engine.
func NewMockProvider() ProviderStrategy {
	return &MockProvider{
		logger: nil, // Will be lazily initialized
	}
}

// WithLogger sets a custom logger for the provider.
// This method enables dependency injection of loggers and allows users to
// configure logging behavior (level, format, output destination).
//
// If not called, a default logger is created automatically on first fetch.
//
// Parameters:
//   - logger: Configured logger instance (must not be nil)
//
// Returns the provider for method chaining.
//
// Example:
//
//	logger := logging.NewLogger(logging.Config{
//	    Level:     logging.LevelDebug,
//	    Format:    logging.FormatJSON,
//	    Component: "provider.mock",
//	})
//
//	provider := provider.NewMockProvider().WithLogger(logger)
func (m *MockProvider) WithLogger(logger *logging.Logger) *MockProvider {
	m.logger = logger
	return m
}

// getLogger returns the provider's logger, creating a default one if needed.
// This implements lazy initialization of the logger to avoid requiring explicit
// configuration in simple use cases.
//
// The default logger uses:
//   - Level: Info
//   - Format: JSON
//   - Component: "provider.mock"
//   - Output: stderr
func (m *MockProvider) getLogger() *logging.Logger {
	if m.logger == nil {
		m.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "provider.mock",
		})
	}
	return m.logger
}

// Fetch retrieves hardcoded sample data for testing and demonstration purposes.
// This implementation returns two records with consistent data on every call.
//
// The method includes comprehensive logging:
//   - Fetch operation start (Info level)
//   - Fetch completion with metrics (Info level)
//   - Warning if no records returned (Warn level - though this never happens in mock)
//
// Returns:
//   - []map[string]interface{}: Slice of data records, each as a key-value map
//   - error: Always nil for MockProvider (included for interface compliance)
//
// Thread-safe: Yes. Multiple goroutines can safely call this method concurrently.
//
// Performance: O(1) - Returns a fixed-size hardcoded dataset.
//
// Example return value:
//
//	[]map[string]interface{}{
//	    {"id": 1, "name": "Alice", "score": 95},
//	    {"id": 2, "name": "Bob", "score": 88},
//	}
func (m *MockProvider) Fetch() ([]map[string]interface{}, error) {
	logger := m.getLogger()
	ctx := context.Background()
	startTime := time.Now()

	logger.InfoContext(ctx, "fetch starting",
		"provider_type", "mock",
		"data_source", "hardcoded",
	)

	// Hardcoded sample data for testing and examples
	data := []map[string]interface{}{
		{
			"id":    1,
			"name":  "Alice",
			"score": 95,
		},
		{
			"id":    2,
			"name":  "Bob",
			"score": 88,
		},
	}

	duration := time.Since(startTime)
	recordCount := len(data)

	logger.InfoContext(ctx, "fetch completed",
		"provider_type", "mock",
		"duration_ms", duration.Milliseconds(),
		"duration_us", duration.Microseconds(),
		"record_count", recordCount,
		"data_source", "hardcoded",
	)

	// Log warning if no records (though this never happens for mock provider)
	if recordCount == 0 {
		logger.WarnContext(ctx, "provider returned zero records",
			"provider_type", "mock",
		)
	}

	// Log debug information about the data structure
	logger.Debug("fetch data structure",
		"provider_type", "mock",
		"fields", []string{"id", "name", "score"},
		"sample_record", data[0],
	)

	return data, nil
}
