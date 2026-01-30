package resilience

import (
	"context"

	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// ProviderWithCircuitBreaker wraps a ProviderStrategy with a CircuitBreaker.
type ProviderWithCircuitBreaker struct {
	delegate provider.ProviderStrategy
	breaker  *CircuitBreaker
}

// NewProviderWithCircuitBreaker creates a new decorator.
func NewProviderWithCircuitBreaker(delegate provider.ProviderStrategy, breaker *CircuitBreaker) *ProviderWithCircuitBreaker {
	return &ProviderWithCircuitBreaker{
		delegate: delegate,
		breaker:  breaker,
	}
}

func (p *ProviderWithCircuitBreaker) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	op := func() error {
		var err error
		results, err = p.delegate.Fetch(ctx)
		return err
	}

	err := p.breaker.Execute(op)
	return results, err
}

func (p *ProviderWithCircuitBreaker) Stream(ctx context.Context) (provider.Iterator, error) {
	if streamer, ok := p.delegate.(provider.StreamingProviderStrategy); ok {
		var iter provider.Iterator
		// Only wrap the initialization logic with circuit breaker
		op := func() error {
			var err error
			iter, err = streamer.Stream(ctx)
			return err
		}

		err := p.breaker.Execute(op)
		return iter, err
	}
	return nil, nil // Or unsupported error
}

func (p *ProviderWithCircuitBreaker) Close() error {
	if closer, ok := p.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}

// OutputWithCircuitBreaker wraps an OutputStrategy with a CircuitBreaker.
type OutputWithCircuitBreaker struct {
	delegate output.OutputStrategy
	breaker  *CircuitBreaker
}

// NewOutputWithCircuitBreaker creates a new decorator.
func NewOutputWithCircuitBreaker(delegate output.OutputStrategy, breaker *CircuitBreaker) *OutputWithCircuitBreaker {
	return &OutputWithCircuitBreaker{
		delegate: delegate,
		breaker:  breaker,
	}
}

func (o *OutputWithCircuitBreaker) Send(ctx context.Context, data []byte) error {
	op := func() error {
		return o.delegate.Send(ctx, data)
	}
	return o.breaker.Execute(op)
}

func (o *OutputWithCircuitBreaker) Close() error {
	if closer, ok := o.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}
