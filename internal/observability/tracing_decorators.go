package observability

import (
	"context"
	"fmt"

	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// ProviderWithTracing wraps a ProviderStrategy with tracing.
type ProviderWithTracing struct {
	delegate provider.ProviderStrategy
	tracer   Tracer
}

func NewProviderWithTracing(delegate provider.ProviderStrategy, tracer Tracer) *ProviderWithTracing {
	return &ProviderWithTracing{
		delegate: delegate,
		tracer:   tracer,
	}
}

func (p *ProviderWithTracing) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	ctx, span := p.tracer.StartSpan(ctx, "provider.fetch")
	defer span.End()

	results, err := p.delegate.Fetch(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetTag("record_count", fmt.Sprintf("%d", len(results)))
	return results, err
}

func (p *ProviderWithTracing) Stream(ctx context.Context) (provider.Iterator, error) {
	if streamer, ok := p.delegate.(provider.StreamingProviderStrategy); ok {
		ctx, span := p.tracer.StartSpan(ctx, "provider.stream_init")
		defer span.End()

		iter, err := streamer.Stream(ctx)
		if err != nil {
			span.RecordError(err)
		}
		return iter, err
	}
	return nil, nil
}

func (p *ProviderWithTracing) Close() error {
	if closer, ok := p.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}

// ProcessorWithTracing wraps a ProcessorHandler with tracing.
type ProcessorWithTracing struct {
	delegate processor.ProcessorHandler
	tracer   Tracer
}

func NewProcessorWithTracing(delegate processor.ProcessorHandler, tracer Tracer) *ProcessorWithTracing {
	return &ProcessorWithTracing{
		delegate: delegate,
		tracer:   tracer,
	}
}

func (p *ProcessorWithTracing) SetNext(next processor.ProcessorHandler) {
	p.delegate.SetNext(next)
}

func (p *ProcessorWithTracing) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	ctx, span := p.tracer.StartSpan(ctx, "processor.process")
	defer span.End()

	span.SetTag("input_count", fmt.Sprintf("%d", len(data)))

	results, err := p.delegate.Process(ctx, data)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetTag("output_count", fmt.Sprintf("%d", len(results)))
	return results, nil
}

// OutputWithTracing wraps an OutputStrategy with tracing.
type OutputWithTracing struct {
	delegate output.OutputStrategy
	tracer   Tracer
}

func NewOutputWithTracing(delegate output.OutputStrategy, tracer Tracer) *OutputWithTracing {
	return &OutputWithTracing{
		delegate: delegate,
		tracer:   tracer,
	}
}

func (o *OutputWithTracing) Send(ctx context.Context, data []byte) error {
	ctx, span := o.tracer.StartSpan(ctx, "output.send")
	defer span.End()

	span.SetTag("bytes_count", fmt.Sprintf("%d", len(data)))

	err := o.delegate.Send(ctx, data)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (o *OutputWithTracing) Close() error {
	if closer, ok := o.delegate.(api.Closeable); ok {
		return closer.Close()
	}
	return nil
}
