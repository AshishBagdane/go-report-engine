package output

import (
	"context"
	"fmt"
)

// ConsoleOutput writes formatted data to stdout.
// It's the simplest output strategy, useful for testing and debugging.
//
// Context handling:
//   - Checks context before writing
//   - Returns ctx.Err() if canceled
//   - Actual write is fast (synchronous to stdout)
//
// Thread-safe: Yes. Multiple goroutines can safely use the same
// ConsoleOutput instance, though output may be interleaved.
type ConsoleOutput struct{}

// NewConsoleOutput creates a new ConsoleOutput instance.
// This is the recommended way to create a ConsoleOutput.
//
// Example:
//
//	output := output.NewConsoleOutput()
//	err := output.Send(ctx, jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Returns:
//   - *ConsoleOutput: A new output instance
func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{}
}

// Send writes data to stdout followed by a newline.
// The data is written as-is without additional formatting.
//
// Context handling:
//   - Checks context before writing
//   - Returns ctx.Err() if already canceled
//   - Write operation is atomic and fast
//
// Note: The actual write to stdout is not interruptible once started,
// but this is typically a fast operation. For writes to slower destinations
// (files, network), implementations should check context more frequently.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - data: Formatted data to write to stdout
//
// Returns:
//   - error: ctx.Err() if context canceled, or write error
func (c *ConsoleOutput) Send(ctx context.Context, data []byte) error {
	// Check if context is already canceled
	// Provides fast path before I/O operation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Write data to stdout
	// fmt.Println adds a newline automatically
	_, err := fmt.Println(string(data))
	return err
}
