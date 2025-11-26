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
