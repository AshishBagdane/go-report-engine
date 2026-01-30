package observability

import (
	"context"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/processor"
)

// ProcessorWithMetrics wraps a ProcessorHandler with metrics collection.
type ProcessorWithMetrics struct {
	delegate  processor.ProcessorHandler
	collector MetricsCollector
}

// NewProcessorWithMetrics creates a new ProcessorWithMetrics decorator.
func NewProcessorWithMetrics(delegate processor.ProcessorHandler, collector MetricsCollector) *ProcessorWithMetrics {
	return &ProcessorWithMetrics{
		delegate:  delegate,
		collector: collector,
	}
}

// SetNext delegates to the underlying processor.
func (p *ProcessorWithMetrics) SetNext(next processor.ProcessorHandler) {
	p.delegate.SetNext(next)
}

// Process executes the delegate's Process method and records metrics.
func (p *ProcessorWithMetrics) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	start := time.Now()
	results, err := p.delegate.Process(ctx, data)
	duration := time.Since(start).Seconds()

	tags := map[string]string{
		"component": "processor",
		"operation": "process",
	}

	p.collector.Histogram("report_engine_processor_duration_seconds", duration, tags)
	p.collector.Count("report_engine_processor_input_records_count", len(data), tags)

	if err != nil {
		p.collector.Count("report_engine_processor_errors_total", 1, tags)
		return nil, err
	}

	p.collector.Count("report_engine_processor_output_records_count", len(results), tags)
	return results, nil
}
