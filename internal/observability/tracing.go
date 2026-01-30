package observability

import "context"

// Tracer defines the interface for creating spans.
// It abstracts away the underlying tracing system (OpenTelemetry, Zipkin, etc.).
type Tracer interface {
	// StartSpan creates a new span and returns a context containing the span.
	// The caller is responsible for calling End() on the returned Span.
	StartSpan(ctx context.Context, name string) (context.Context, Span)
}

// Span represents a single operation within a trace.
type Span interface {
	// End marks the end of the span execution.
	End()

	// SetTag adds a key-value pair to the span.
	SetTag(key, value string)

	// RecordError records an error on the span.
	RecordError(err error)
}

// NoopTracer is a default implementation that does nothing.
type NoopTracer struct{}

func (n *NoopTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

// NoopSpan is a default implementation that does nothing.
type NoopSpan struct{}

func (n *NoopSpan) End()                     {}
func (n *NoopSpan) SetTag(key, value string) {}
func (n *NoopSpan) RecordError(err error)    {}

// NewNoopTracer creates a new NoopTracer.
func NewNoopTracer() Tracer {
	return &NoopTracer{}
}
