package resilience

import (
	"context"

	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// OutputWithRetry wraps an OutputStrategy with retry logic.
type OutputWithRetry struct {
	delegate output.OutputStrategy
	retrier  *Retrier
}

// NewOutputWithRetry creates a new OutputWithRetry decorator.
func NewOutputWithRetry(delegate output.OutputStrategy, retrier *Retrier) *OutputWithRetry {
	return &OutputWithRetry{
		delegate: delegate,
		retrier:  retrier,
	}
}

// Send executes the delegate's Send method with retries.
func (o *OutputWithRetry) Send(ctx context.Context, data []byte) error {
	op := func(ctx context.Context) error {
		return o.delegate.Send(ctx, data)
	}

	return o.retrier.Execute(ctx, op)
}

// Close delegates to the underlying output if it implements Closeable.
func (o *OutputWithRetry) Close() error {
	if closer, ok := o.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}
