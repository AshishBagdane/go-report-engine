package formatter

import "context"

// FormatStrategy defines the interface for data formatters.
// All formatters must implement this interface to participate in the
// report generation pipeline.
//
// The context parameter enables:
//   - Cancellation of long-running formatting operations
//   - Timeout enforcement for large datasets
//   - Propagation of cancellation signals
//   - Request tracing via context values
//
// Implementations MUST:
//   - Check ctx.Done() before starting expensive operations
//   - Check ctx.Done() periodically during large iterations
//   - Return ctx.Err() when context is canceled
//   - Handle partial formatting gracefully
//   - Clean up resources when canceled
//
// Thread-safety: Formatters may be called concurrently if used in
// multiple engine instances. Implementations must be thread-safe if shared.
type FormatStrategy interface {
	// Format converts processed data into a specific output format.
	// The context allows for cancellation and timeout control.
	//
	// The formatter receives structured data (slice of maps) and
	// produces a byte slice representing the formatted output.
	//
	// For large datasets, implementations should:
	//   - Check ctx.Done() periodically
	//   - Use efficient serialization libraries
	//   - Consider streaming approaches
	//   - Pre-allocate buffers when possible
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and tracing
	//   - data: Processed data to format
	//
	// Returns:
	//   - []byte: The formatted output
	//   - error: Any error that occurred, including:
	//     - context.Canceled if ctx was canceled
	//     - context.DeadlineExceeded if ctx deadline exceeded
	//     - Serialization errors specific to the format
	//
	// Implementations should return promptly when ctx.Done() is closed.
	Format(ctx context.Context, data []map[string]interface{}) ([]byte, error)
}
