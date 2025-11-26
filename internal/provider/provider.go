package provider

import "context"

// ProviderStrategy defines the interface for data providers.
// All providers must implement this interface to participate in the
// report generation pipeline.
//
// The context parameter enables:
//   - Cancellation of long-running fetch operations
//   - Timeout enforcement via context.WithTimeout
//   - Deadline propagation via context.WithDeadline
//   - Request tracing via context values (request ID, correlation ID)
//
// Implementations MUST:
//   - Check ctx.Done() periodically during long operations
//   - Return ctx.Err() when context is canceled or deadline exceeded
//   - Pass context to any underlying I/O operations
//   - Clean up resources when context is canceled
//
// Thread-safety: Implementations may be called concurrently from multiple
// goroutines and must be thread-safe if shared across goroutines.
type ProviderStrategy interface {
	// Fetch retrieves data from the provider's source.
	// The context allows for cancellation and timeout control.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and tracing
	//
	// Returns:
	//   - []map[string]interface{}: The fetched data records
	//   - error: Any error that occurred during fetch, including:
	//     - context.Canceled if ctx was canceled
	//     - context.DeadlineExceeded if ctx deadline was exceeded
	//     - Provider-specific errors for data source issues
	//
	// Implementations should return promptly when ctx.Done() is closed.
	Fetch(ctx context.Context) ([]map[string]interface{}, error)
}
