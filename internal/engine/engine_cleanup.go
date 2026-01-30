package engine

import (
	"context"
	"io"
	"time"

	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// Close implements io.Closer for resource cleanup.
// It closes all components in reverse order (LIFO) that implement cleanup interfaces.
//
// Cleanup order (reverse of construction):
//  1. Output
//  2. Formatter
//  3. Processor
//  4. Provider
//
// This order ensures that components depending on others are closed first.
//
// The method is idempotent - calling Close() multiple times is safe.
// Subsequent calls after the first will return nil.
//
// Thread-safe: Yes. Multiple goroutines can safely call Close() concurrently.
//
// Example:
//
//	engine, err := factory.NewEngineFromConfig(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer engine.Close()
//
//	if err := engine.Run(); err != nil {
//	    log.Printf("Pipeline failed: %v", err)
//	}
//
// Returns:
//   - error: Aggregated error if any cleanup failed, or nil if all succeeded
func (r *ReportEngine) Close() error {
	return r.CloseWithContext(context.Background())
}

// CloseWithContext performs context-aware cleanup of all components.
// It supports graceful shutdown with timeout control.
//
// The method closes all components in reverse order (LIFO) that implement
// cleanup interfaces. It first attempts CloseableWithContext, then falls
// back to Closeable.
//
// Cleanup order (reverse of construction):
//  1. Output
//  2. Formatter
//  3. Processor
//  4. Provider
//
// The method is idempotent - calling CloseWithContext() multiple times is safe.
// Subsequent calls after the first will return nil.
//
// Thread-safe: Yes. Multiple goroutines can safely call CloseWithContext() concurrently.
//
// Parameters:
//   - ctx: Context for timeout and cancellation control
//
// Returns:
//   - error: Aggregated error if any cleanup failed, including:
//   - context.Canceled if ctx was canceled during cleanup
//   - context.DeadlineExceeded if cleanup exceeded deadline
//   - Component-specific cleanup errors
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if err := engine.CloseWithContext(ctx); err != nil {
//	    log.Printf("Cleanup failed: %v", err)
//	}
func (r *ReportEngine) CloseWithContext(ctx context.Context) error {
	// Use sync.Once to ensure cleanup happens only once
	var closeErr error

	r.closeOnce.Do(func() {
		var closer api.MultiCloser

		// Add components in reverse order (LIFO)
		// Output -> Formatter -> Processor -> Provider

		// Close Output
		if r.Output != nil {
			if c, ok := r.Output.(api.CloseableWithContext); ok {
				// Wrap CloseableWithContext to match io.Closer interface
				closer.Add(&contextCloserAdapter{c: c, ctx: ctx})
			} else if c, ok := r.Output.(io.Closer); ok {
				closer.Add(c)
			}
		}

		// Close Formatter
		if r.Formatter != nil {
			if c, ok := r.Formatter.(api.CloseableWithContext); ok {
				closer.Add(&contextCloserAdapter{c: c, ctx: ctx})
			} else if c, ok := r.Formatter.(io.Closer); ok {
				closer.Add(c)
			}
		}

		// Close Processor
		if r.Processor != nil {
			if c, ok := r.Processor.(api.CloseableWithContext); ok {
				closer.Add(&contextCloserAdapter{c: c, ctx: ctx})
			} else if c, ok := r.Processor.(io.Closer); ok {
				closer.Add(c)
			}
		}

		// Close Provider
		if r.Provider != nil {
			if c, ok := r.Provider.(api.CloseableWithContext); ok {
				closer.Add(&contextCloserAdapter{c: c, ctx: ctx})
			} else if c, ok := r.Provider.(io.Closer); ok {
				closer.Add(c)
			}
		}

		// Execute cleanup
		closeErr = closer.Close()

		// Log cleanup completion
		logger := r.getLogger()
		if closeErr != nil {
			logger.Error("engine cleanup completed with errors",
				"error", closeErr,
			)
		} else {
			logger.Info("engine cleanup completed successfully")
		}
	})

	return closeErr
}

// Shutdown performs graceful shutdown with a timeout.
// It's a convenience wrapper around CloseWithContext that creates
// a timeout context automatically.
//
// This method is useful for server shutdown handlers where you want
// to enforce a maximum cleanup time.
//
// Parameters:
//   - timeout: Maximum time to wait for cleanup
//
// Returns:
//   - error: Aggregated error if any cleanup failed, or timeout error
//
// Example:
//
//	// In server shutdown handler
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
//	<-sigChan
//
//	log.Println("Shutting down gracefully...")
//	if err := engine.Shutdown(30 * time.Second); err != nil {
//	    log.Printf("Shutdown failed: %v", err)
//	}
func (r *ReportEngine) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return r.CloseWithContext(ctx)
}

// contextCloserAdapter adapts CloseableWithContext to io.Closer interface.
// It allows CloseableWithContext implementations to be used with MultiCloser.
type contextCloserAdapter struct {
	c   api.CloseableWithContext
	ctx context.Context
}

// Close implements io.Closer by calling CloseWithContext.
func (a *contextCloserAdapter) Close() error {
	return a.c.CloseWithContext(a.ctx)
}
