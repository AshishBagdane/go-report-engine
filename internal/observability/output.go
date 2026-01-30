package observability

import (
	"context"
	"time"

	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// OutputWithMetrics wraps an OutputStrategy with metrics collection.
type OutputWithMetrics struct {
	delegate  output.OutputStrategy
	collector MetricsCollector
}

// NewOutputWithMetrics creates a new OutputWithMetrics decorator.
func NewOutputWithMetrics(delegate output.OutputStrategy, collector MetricsCollector) *OutputWithMetrics {
	return &OutputWithMetrics{
		delegate:  delegate,
		collector: collector,
	}
}

// Send executes the delegate's Send method and records metrics.
func (o *OutputWithMetrics) Send(ctx context.Context, data []byte) error {
	start := time.Now()
	err := o.delegate.Send(ctx, data)
	duration := time.Since(start).Seconds()

	tags := map[string]string{
		"component": "output",
		"operation": "send",
	}

	o.collector.Histogram("report_engine_output_duration_seconds", duration, tags)
	o.collector.Count("report_engine_output_bytes_count", len(data), tags)

	if err != nil {
		o.collector.Count("report_engine_output_errors_total", 1, tags)
		return err
	}

	return nil
}

// Close delegates to the underlying output if it implements Closeable.
func (o *OutputWithMetrics) Close() error {
	if closer, ok := o.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}
