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
//
// Resource Cleanup (Optional):
// Providers that manage resources (database connections, file handles, network
// connections, etc.) SHOULD implement one of the cleanup interfaces from pkg/api:
//   - api.Closeable - for simple cleanup
//   - api.CloseableWithContext - for context-aware cleanup with timeout support
//
// The engine will automatically detect and call cleanup methods via type assertion.
// Providers that don't manage resources don't need to implement cleanup.
//
// Example provider with cleanup:
//
//	type DatabaseProvider struct {
//	    pool      *sql.DB
//	    closeOnce sync.Once
//	    closeErr  error
//	}
//
//	func (d *DatabaseProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
//	    // Fetch data from database pool
//	    return data, nil
//	}
//
//	func (d *DatabaseProvider) Close() error {
//	    d.closeOnce.Do(func() {
//	        if d.pool != nil {
//	            d.closeErr = d.pool.Close()
//	        }
//	    })
//	    return d.closeErr
//	}
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

// Iterator defines the interface for iterating over data records one by one.
// This enables memory-efficient processing of large datasets.
type Iterator interface {
	// Next advances the iterator to the next record.
	// Returns true if a record is available, false if the stream is exhausted or an error occurred.
	Next() bool

	// Value returns the current record.
	// Should only be called after a successful Next() call.
	Value() map[string]interface{}

	// Err returns any error that occurred during iteration.
	// Should be checked after Next() returns false.
	Err() error

	// Close releases any resources associated with the iterator.
	Close() error
}

// StreamingProviderStrategy extends ProviderStrategy to support streaming data access.
// Providers that support streaming should implement this interface.
type StreamingProviderStrategy interface {
	ProviderStrategy

	// Stream returns an Iterator for streaming data access.
	// The context controls the lifetime of the stream initialization.
	Stream(ctx context.Context) (Iterator, error)
}
