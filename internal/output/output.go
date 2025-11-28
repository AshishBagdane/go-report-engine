package output

import "context"

// OutputStrategy defines the interface for output delivery.
// All outputs must implement this interface to participate in the
// report generation pipeline.
//
// The context parameter enables:
//   - Cancellation of long-running delivery operations
//   - Timeout enforcement for network operations
//   - Propagation of cancellation signals
//   - Request tracing via context values
//
// Implementations MUST:
//   - Check ctx.Done() before starting I/O operations
//   - Pass context to underlying I/O operations (HTTP, S3, etc.)
//   - Return ctx.Err() when context is canceled
//   - Clean up resources (files, connections) when canceled
//   - Handle partial writes appropriately
//
// Thread-safety: Outputs may be called concurrently if used in
// multiple engine instances. Implementations must be thread-safe if shared.
//
// Resource Cleanup (Optional):
// Outputs that manage resources (file handles, network connections, HTTP clients,
// database connections, message queues, cloud service clients, etc.) SHOULD
// implement one of the cleanup interfaces from pkg/api:
//   - api.Closeable - for simple cleanup
//   - api.CloseableWithContext - for context-aware cleanup with timeout support
//
// The engine will automatically detect and call cleanup methods via type assertion.
// Simple outputs that don't maintain persistent resources (like ConsoleOutput)
// don't need to implement cleanup.
//
// Example output with cleanup:
//
//	type FileOutput struct {
//	    file      *os.File
//	    writer    *bufio.Writer
//	    closeOnce sync.Once
//	    closeErr  error
//	}
//
//	func (f *FileOutput) Send(ctx context.Context, data []byte) error {
//	    select {
//	    case <-ctx.Done():
//	        return ctx.Err()
//	    default:
//	    }
//	    _, err := f.writer.Write(data)
//	    return err
//	}
//
//	func (f *FileOutput) Close() error {
//	    f.closeOnce.Do(func() {
//	        if f.writer != nil {
//	            f.writer.Flush()
//	        }
//	        if f.file != nil {
//	            f.closeErr = f.file.Close()
//	        }
//	    })
//	    return f.closeErr
//	}
type OutputStrategy interface {
	// Send delivers the formatted data to its destination.
	// The context allows for cancellation and timeout control.
	//
	// For I/O operations, implementations should:
	//   - Use context-aware APIs (e.g., http.NewRequestWithContext)
	//   - Check ctx.Done() for long-running operations
	//   - Handle partial writes gracefully
	//   - Clean up resources on cancellation
	//
	// Common patterns:
	//   - Network: Use request.WithContext(ctx)
	//   - File I/O: Write in chunks, check ctx.Done() between chunks
	//   - Cloud APIs: Pass context to SDK methods
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and tracing
	//   - data: Formatted data to deliver
	//
	// Returns:
	//   - error: Any error that occurred, including:
	//     - context.Canceled if ctx was canceled
	//     - context.DeadlineExceeded if ctx deadline exceeded
	//     - I/O errors specific to the output destination
	//
	// Implementations should return promptly when ctx.Done() is closed.
	Send(ctx context.Context, data []byte) error
}
