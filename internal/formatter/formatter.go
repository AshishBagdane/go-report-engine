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
//
// Resource Cleanup (Optional):
// Formatters that manage resources (output buffers, file handles, streaming
// encoders, compression streams, etc.) SHOULD implement one of the cleanup
// interfaces from pkg/api:
//   - api.Closeable - for simple cleanup
//   - api.CloseableWithContext - for context-aware cleanup with timeout support
//
// The engine will automatically detect and call cleanup methods via type assertion.
// Stateless formatters that only serialize in-memory don't need to implement cleanup.
//
// Example formatter with cleanup:
//
//	type StreamingFormatter struct {
//	    writer    *bufio.Writer
//	    file      *os.File
//	    closeOnce sync.Once
//	    closeErr  error
//	}
//
//	func (s *StreamingFormatter) Format(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
//	    // Stream formatting to buffer
//	    return result, nil
//	}
//
//	func (s *StreamingFormatter) Close() error {
//	    s.closeOnce.Do(func() {
//	        if s.writer != nil {
//	            s.writer.Flush()
//	        }
//	        if s.file != nil {
//	            s.closeErr = s.file.Close()
//	        }
//	    })
//	    return s.closeErr
//	}
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
