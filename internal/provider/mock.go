package provider

import (
	"context"
	"time"

	"github.com/AshishBagdane/report-engine/internal/logging"
)

// MockProvider is a simple in-memory provider for testing and examples.
// It returns pre-configured data without accessing external sources.
//
// MockProvider respects context cancellation even though its operation
// is fast. This demonstrates proper context handling patterns.
//
// Thread-safe: Yes. MockProvider is immutable after creation and can
// be safely called from multiple goroutines.
type MockProvider struct {
	// Data is the pre-configured data to return
	Data []map[string]interface{}

	// logger is the optional logger instance for this provider
	logger *logging.Logger
}

// NewMockProvider creates a new MockProvider with the given data.
// This is the recommended way to create a MockProvider.
//
// Example:
//
//	provider := provider.NewMockProvider([]map[string]interface{}{
//	    {"id": 1, "name": "Alice", "score": 95},
//	    {"id": 2, "name": "Bob", "score": 87},
//	})
//
// Parameters:
//   - data: The data to return from Fetch calls
//
// Returns:
//   - *MockProvider: A new provider instance
func NewMockProvider(data []map[string]interface{}) *MockProvider {
	return &MockProvider{
		Data:   data,
		logger: nil, // Will be lazily initialized if needed
	}
}

// WithLogger sets the logger for this provider and returns the provider for chaining.
// If no logger is set, a default logger will be created on first use.
//
// Example:
//
//	logger := logging.NewLogger(logging.Config{
//	    Level:     logging.LevelInfo,
//	    Component: "provider.mock",
//	})
//	provider := provider.NewMockProvider(data).WithLogger(logger)
func (m *MockProvider) WithLogger(logger *logging.Logger) *MockProvider {
	m.logger = logger
	return m
}

// getLogger returns the logger instance, creating a default one if necessary.
// This ensures lazy initialization of the logger.
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

// Fetch returns the pre-configured data from the provider.
// It checks for context cancellation before returning to demonstrate
// proper context handling, even for fast operations.
//
// Context handling:
//   - Returns ctx.Err() if context is already canceled/expired
//   - Returns immediately if context is valid
//
// Logging:
//   - Logs fetch start and completion at Info level
//   - Logs data structure details at Debug level
//   - Includes performance metrics (duration, record count)
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - []map[string]interface{}: The mock data
//   - error: ctx.Err() if context is canceled/expired, nil otherwise
func (m *MockProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	logger := m.getLogger()
	startTime := time.Now()

	// Log fetch starting
	logger.InfoContext(ctx, "fetch starting",
		"provider_type", "mock",
		"data_size", len(m.Data),
	)

	// Check if context is already canceled or deadline exceeded
	// This is important even for fast operations to respect cancellation
	select {
	case <-ctx.Done():
		// Context was canceled or deadline exceeded
		logger.WarnContext(ctx, "fetch canceled",
			"provider_type", "mock",
			"reason", ctx.Err().Error(),
		)
		return nil, ctx.Err()
	default:
		// Context is still valid, proceed with fetch
	}

	// Log data structure at debug level
	if logger.Enabled(logging.LevelDebug) && len(m.Data) > 0 {
		// Extract field names from first record
		fields := make([]string, 0, len(m.Data[0]))
		for key := range m.Data[0] {
			fields = append(fields, key)
		}
		logger.DebugContext(ctx, "fetch data structure",
			"provider_type", "mock",
			"fields", fields,
			"sample_record", m.Data[0],
		)
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Log fetch completion
	logger.InfoContext(ctx, "fetch completed",
		"provider_type", "mock",
		"record_count", len(m.Data),
		"duration_ms", float64(duration.Microseconds())/1000.0,
	)

	// Return the pre-configured data
	return m.Data, nil
}
