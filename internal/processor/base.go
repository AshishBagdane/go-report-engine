package processor

import "context"

// BaseProcessor is a basic processor that passes data through unchanged.
// It serves as:
//   - A pass-through processor when no transformation is needed
//   - A base struct for embedding in custom processors
//   - The end of a processor chain (if next is nil)
//
// BaseProcessor properly handles context cancellation and propagates
// it through the chain.
//
// Thread-safe: Yes. BaseProcessor operations are thread-safe as long
// as SetNext is not called concurrently with Process.
type BaseProcessor struct {
	// next is the next processor in the chain
	next ProcessorHandler
}

// SetNext sets the next processor in the chain.
// This should typically be called during chain construction,
// not during processing.
//
// Not safe for concurrent calls with Process. Chain construction
// should be completed before processing begins.
//
// Parameters:
//   - next: The next processor to call after this one
func (b *BaseProcessor) SetNext(next ProcessorHandler) {
	b.next = next
}

// Process passes data through to the next processor in the chain.
// If there is no next processor, it returns the data unchanged.
//
// Context handling:
//   - Checks context cancellation before proceeding
//   - Propagates context to next processor
//   - Returns ctx.Err() if context is canceled
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - data: Input data to process
//
// Returns:
//   - []map[string]interface{}: Unmodified data (or result from next processor)
//   - error: ctx.Err() if context is canceled, or error from next processor
func (b *BaseProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check if context is canceled before processing
	// This provides a fast path for cancellation
	select {
	case <-ctx.Done():
		// Context was canceled or deadline exceeded
		return nil, ctx.Err()
	default:
		// Context is still valid, proceed
	}

	// If there's a next processor, pass data through with context
	if b.next != nil {
		return b.next.Process(ctx, data)
	}

	// End of chain - return data unchanged
	return data, nil
}
