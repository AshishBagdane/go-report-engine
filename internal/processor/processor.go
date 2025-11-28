package processor

import "context"

// ProcessorHandler defines the interface for processors in the chain.
// Each processor can perform operations on the data and pass it to the
// next processor in the chain.
//
// The Chain of Responsibility pattern allows for:
//   - Flexible, composable data transformations
//   - Independent testing of each processor
//   - Runtime chain reconfiguration
//   - Short-circuiting on errors
//
// Context support enables:
//   - Cancellation of long-running processing operations
//   - Timeout enforcement for the entire chain
//   - Propagation of cancellation through the chain
//   - Request tracing with request/correlation IDs
//
// Implementations MUST:
//   - Check ctx.Done() periodically during loops
//   - Return ctx.Err() when context is canceled
//   - Pass context to next processor in chain
//   - Handle context cancellation gracefully
//   - Clean up resources when canceled
//
// Thread-safety: Processors may be called concurrently if used in
// multiple engine instances. Stateful processors must be thread-safe.
//
// Resource Cleanup (Optional):
// Processors that manage resources (file buffers, network connections,
// caches, background goroutines, etc.) SHOULD implement one of the cleanup
// interfaces from pkg/api:
//   - api.Closeable - for simple cleanup
//   - api.CloseableWithContext - for context-aware cleanup with timeout support
//
// The engine will automatically detect and call cleanup methods via type assertion.
// Stateless processors don't need to implement cleanup.
//
// For processor chains, cleanup will be called on the head processor, which
// should propagate cleanup through the chain if needed.
//
// Example processor with cleanup:
//
//	type CachingProcessor struct {
//	    BaseProcessor
//	    cache     *Cache
//	    closeOnce sync.Once
//	    closeErr  error
//	}
//
//	func (c *CachingProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
//	    // Process with cache
//	    return c.BaseProcessor.Process(ctx, data)
//	}
//
//	func (c *CachingProcessor) Close() error {
//	    c.closeOnce.Do(func() {
//	        if c.cache != nil {
//	            c.closeErr = c.cache.Close()
//	        }
//	        // Propagate to next processor if it supports cleanup
//	        if c.next != nil {
//	            if closer, ok := c.next.(io.Closer); ok {
//	                if err := closer.Close(); err != nil && c.closeErr == nil {
//	                    c.closeErr = err
//	                }
//	            }
//	        }
//	    })
//	    return c.closeErr
//	}
type ProcessorHandler interface {
	// SetNext sets the next processor in the chain.
	// This allows for runtime chain construction.
	//
	// Parameters:
	//   - next: The next processor to call after this one
	SetNext(next ProcessorHandler)

	// Process transforms the input data and passes it to the next processor.
	// The context enables cancellation and timeout control.
	//
	// Processors should:
	//   1. Check ctx.Done() before starting expensive operations
	//   2. Check ctx.Done() periodically during loops
	//   3. Pass context to next processor in chain
	//   4. Return ctx.Err() when context is canceled
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and tracing
	//   - data: Input data to process
	//
	// Returns:
	//   - []map[string]interface{}: Processed data
	//   - error: Any error that occurred, including:
	//     - context.Canceled if ctx was canceled
	//     - context.DeadlineExceeded if ctx deadline exceeded
	//     - Validation/processing errors specific to the processor
	//
	// The returned error will propagate up the chain and halt processing.
	Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error)
}
