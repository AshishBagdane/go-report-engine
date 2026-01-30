package resilience

import (
	"context"

	"github.com/AshishBagdane/report-engine/internal/provider"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// ProviderWithRetry wraps a ProviderStrategy with retry logic.
type ProviderWithRetry struct {
	delegate provider.ProviderStrategy
	retrier  *Retrier
}

// NewProviderWithRetry creates a new ProviderWithRetry decorator.
func NewProviderWithRetry(delegate provider.ProviderStrategy, retrier *Retrier) *ProviderWithRetry {
	return &ProviderWithRetry{
		delegate: delegate,
		retrier:  retrier,
	}
}

// Fetch executes the delegate's Fetch method with retries.
func (p *ProviderWithRetry) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	op := func(ctx context.Context) error {
		var err error
		results, err = p.delegate.Fetch(ctx)
		return err
	}

	err := p.retrier.Execute(ctx, op)
	return results, err
}

// Stream executes the delegate's Stream method **without** retries for the stream itself.
func (p *ProviderWithRetry) Stream(ctx context.Context) (provider.Iterator, error) {
	if streamer, ok := p.delegate.(provider.StreamingProviderStrategy); ok {
		return streamer.Stream(ctx)
	}
	// Return nil if not supported, or wrapped error?
	// Usually the caller checks for interface compliance before calling, or we assume they know.
	// But if we return nil iterator, caller might panic.
	// Ideally we only expose Stream if delegate does.
	// But struct embedding isn't dynamic.
	return nil, nil // TODO: Proper error "Streaming not supported" or similar.
}

// Close delegates cleanup if supported.
func (p *ProviderWithRetry) Close() error {
	if closer, ok := p.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}
