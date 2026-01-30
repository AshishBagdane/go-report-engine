// Package api defines the public interfaces and types for the report engine.
// This file defines lifecycle management interfaces for proper resource cleanup.
package api

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Closeable extends io.Closer and represents a component that owns resources
// requiring explicit cleanup. Components implementing this interface will have
// their Close() method called during engine shutdown.
//
// This interface is optional - only components that manage resources like
// file handles, database connections, network connections, or other system
// resources need to implement it.
//
// Implementations MUST:
//   - Be idempotent (safe to call Close() multiple times)
//   - Release all resources (file handles, connections, memory)
//   - NOT panic, even if called multiple times
//   - Return meaningful errors if cleanup fails
//   - Clean up even if the component is in an error state
//
// Implementations SHOULD:
//   - Use sync.Once or flags to prevent double-cleanup
//   - Log cleanup operations for debugging
//   - Aggregate multiple cleanup errors if managing multiple resources
//
// Thread-safety: Close() may be called concurrently with other operations
// if the component is shared across goroutines. Implementations must handle
// this safely, typically by using sync.Once or proper locking.
//
// Example implementation:
//
//	type FileProvider struct {
//	    file      *os.File
//	    closeOnce sync.Once
//	    closeErr  error
//	}
//
//	func (f *FileProvider) Close() error {
//	    f.closeOnce.Do(func() {
//	        if f.file != nil {
//	            f.closeErr = f.file.Close()
//	        }
//	    })
//	    return f.closeErr
//	}
type Closeable interface {
	io.Closer
}

// CloseableWithContext represents a component that supports context-aware cleanup.
// This allows for graceful shutdown with timeout control and cancellation.
//
// This interface is preferred over simple Closeable when cleanup operations might:
//   - Take significant time (flushing buffers, draining queues)
//   - Involve network I/O (closing remote connections gracefully)
//   - Need coordination (waiting for in-flight operations)
//   - Require timeout enforcement (preventing hung shutdowns)
//
// Implementations MUST:
//   - Respect context cancellation and deadline
//   - Return ctx.Err() if context expires during cleanup
//   - Be idempotent (safe to call multiple times)
//   - Handle partial cleanup gracefully
//   - Still clean up critical resources even if context is canceled
//
// Implementations SHOULD:
//   - Attempt graceful cleanup first (flush, drain, close gracefully)
//   - Fall back to forceful cleanup if context is canceled
//   - Check ctx.Done() periodically during long operations
//   - Use context.WithTimeout internally if needed
//
// Thread-safety: CloseWithContext() may be called concurrently with other
// operations. Implementations must handle this safely.
//
// Example implementation:
//
//	type DatabaseProvider struct {
//	    pool      *sql.DB
//	    closeOnce sync.Once
//	    closeErr  error
//	}
//
//	func (d *DatabaseProvider) CloseWithContext(ctx context.Context) error {
//	    d.closeOnce.Do(func() {
//	        if d.pool == nil {
//	            return
//	        }
//
//	        // Create a channel for cleanup completion
//	        done := make(chan error, 1)
//
//	        go func() {
//	            // Attempt graceful connection pool shutdown
//	            done <- d.pool.Close()
//	        }()
//
//	        // Wait for cleanup or context cancellation
//	        select {
//	        case err := <-done:
//	            d.closeErr = err
//	        case <-ctx.Done():
//	            // Context canceled, but we still need to cleanup
//	            // Fall back to forceful close if needed
//	            d.closeErr = ctx.Err()
//	        }
//	    })
//	    return d.closeErr
//	}
type CloseableWithContext interface {
	// CloseWithContext performs context-aware cleanup of resources.
	//
	// The context allows the caller to:
	//   - Set a deadline for cleanup operations
	//   - Cancel cleanup if taking too long
	//   - Propagate cancellation signals
	//
	// The method should attempt graceful cleanup within the context deadline,
	// but must still release critical resources even if context is canceled.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation control
	//
	// Returns:
	//   - error: Any error during cleanup, including:
	//     - context.Canceled if ctx was canceled during cleanup
	//     - context.DeadlineExceeded if cleanup exceeded deadline
	//     - Resource-specific errors (connection close failed, etc.)
	CloseWithContext(ctx context.Context) error
}

// MultiCloser aggregates multiple Closeable components and provides
// coordinated cleanup. It's useful when a component owns multiple
// sub-resources that each need cleanup.
//
// MultiCloser closes all components in reverse order (LIFO) and aggregates
// all errors. This ensures that even if one component fails to close,
// others are still attempted.
//
// Thread-safety: MultiCloser is NOT thread-safe. The caller must ensure
// that Add() is not called concurrently with Close().
//
// Example usage:
//
//	type ComplexComponent struct {
//	    closer MultiCloser
//	    file   *os.File
//	    conn   net.Conn
//	}
//
//	func NewComplexComponent() (*ComplexComponent, error) {
//	    c := &ComplexComponent{}
//
//	    file, err := os.Open("data.txt")
//	    if err != nil {
//	        return nil, err
//	    }
//	    c.file = file
//	    c.closer.Add(file)
//
//	    conn, err := net.Dial("tcp", "localhost:8080")
//	    if err != nil {
//	        c.closer.Close() // Clean up file
//	        return nil, err
//	    }
//	    c.conn = conn
//	    c.closer.Add(conn)
//
//	    return c, nil
//	}
//
//	func (c *ComplexComponent) Close() error {
//	    return c.closer.Close()
//	}
type MultiCloser struct {
	mu      sync.Mutex
	closers []io.Closer
}

// Add registers a Closeable component for cleanup.
// Components are closed in reverse order (LIFO - last in, first out).
//
// This method is NOT thread-safe and should only be called during
// initialization, not concurrently with Close().
//
// Parameters:
//   - closer: Component to register for cleanup (must not be nil)
//
// Example:
//
//	var mc MultiCloser
//	mc.Add(file)
//	mc.Add(connection)
//	defer mc.Close() // Closes connection, then file
func (m *MultiCloser) Add(closer io.Closer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if closer != nil {
		m.closers = append(m.closers, closer)
	}
}

// Close closes all registered components in reverse order (LIFO).
// It attempts to close all components even if some fail, and aggregates
// all errors into a single error.
//
// The method is idempotent - subsequent calls after the first Close()
// will return nil since the closers slice is cleared.
//
// Returns:
//   - error: Aggregated error if any cleanup failed, or nil if all succeeded
//
// Example:
//
//	err := mc.Close()
//	if err != nil {
//	    log.Printf("Cleanup failed: %v", err)
//	}
func (m *MultiCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	// Close in reverse order (LIFO)
	for i := len(m.closers) - 1; i >= 0; i-- {
		if err := m.closers[i].Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// Clear the slice to make Close() idempotent
	m.closers = nil

	// Return aggregated errors
	if len(errs) > 0 {
		return &CloseErrors{Errors: errs}
	}

	return nil
}

// CloseErrors aggregates multiple errors that occurred during cleanup.
// It implements the error interface and provides a formatted message
// showing all cleanup failures.
//
// Example error message:
//
//	"cleanup failed with 3 errors: [error1, error2, error3]"
type CloseErrors struct {
	// Errors contains all errors that occurred during cleanup
	Errors []error
}

// Error implements the error interface.
// It formats all errors into a single message for logging and debugging.
//
// Returns:
//   - string: Formatted error message with all cleanup failures
func (e *CloseErrors) Error() string {
	if len(e.Errors) == 0 {
		return "cleanup completed with no errors"
	}

	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	msg := fmt.Sprintf("cleanup failed with %d errors: [", len(e.Errors))
	for i, err := range e.Errors {
		if i > 0 {
			msg += ", "
		}
		msg += err.Error()
	}
	msg += "]"

	return msg
}

// Unwrap returns the underlying errors for error inspection.
// This supports Go 1.20+ multi-error unwrapping.
//
// Returns:
//   - []error: All cleanup errors
func (e *CloseErrors) Unwrap() []error {
	return e.Errors
}
