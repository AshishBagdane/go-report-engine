package observability

import (
	"context"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// ProviderWithMetrics wraps a ProviderStrategy with metrics collection.
type ProviderWithMetrics struct {
	delegate  provider.ProviderStrategy
	collector MetricsCollector
}

// NewProviderWithMetrics creates a new ProviderWithMetrics decorator.
func NewProviderWithMetrics(delegate provider.ProviderStrategy, collector MetricsCollector) *ProviderWithMetrics {
	return &ProviderWithMetrics{
		delegate:  delegate,
		collector: collector,
	}
}

// Fetch executes the delegate's Fetch method and records metrics.
func (p *ProviderWithMetrics) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	start := time.Now()
	results, err := p.delegate.Fetch(ctx)
	duration := time.Since(start).Seconds()

	tags := map[string]string{
		"component": "provider",
		"operation": "fetch",
	}

	p.collector.Histogram("report_engine_provider_fetch_duration_seconds", duration, tags)

	if err != nil {
		p.collector.Count("report_engine_provider_errors_total", 1, tags)
		return nil, err
	}

	p.collector.Count("report_engine_provider_records_count", len(results), tags)
	return results, nil
}

// Stream wraps the streaming iterator with metrics if supported.
func (p *ProviderWithMetrics) Stream(ctx context.Context) (provider.Iterator, error) {
	if streamer, ok := p.delegate.(provider.StreamingProviderStrategy); ok {
		start := time.Now()
		iter, err := streamer.Stream(ctx)
		duration := time.Since(start).Seconds()

		tags := map[string]string{
			"component": "provider",
			"operation": "stream_init",
		}
		p.collector.Histogram("report_engine_provider_stream_init_duration_seconds", duration, tags)

		if err != nil {
			p.collector.Count("report_engine_provider_errors_total", 1, tags)
		}

		return iter, err
	}
	return nil, nil
}
