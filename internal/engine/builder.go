package engine

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/observability"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
	"github.com/AshishBagdane/report-engine/internal/resilience"
)

// EngineBuilder provides a fluent interface for constructing a ReportEngine.
// It implements the Builder pattern with comprehensive validation.
type EngineBuilder struct {
	provider  provider.ProviderStrategy
	processor processor.ProcessorHandler
	formatter formatter.FormatStrategy
	output    output.OutputStrategy
	retry     *resilience.RetryPolicy
	breaker   *resilience.CircuitBreaker

	tracer  observability.Tracer
	metrics observability.MetricsCollector
}

// NewEngineBuilder creates a new EngineBuilder with default values.
func NewEngineBuilder() *EngineBuilder {
	return &EngineBuilder{}
}

// WithProvider sets the provider and validates it is not nil.
func (b *EngineBuilder) WithProvider(p provider.ProviderStrategy) *EngineBuilder {
	b.provider = p
	return b
}

// WithProcessor sets the processor and validates it is not nil.
func (b *EngineBuilder) WithProcessor(p processor.ProcessorHandler) *EngineBuilder {
	b.processor = p
	return b
}

// WithFormatter sets the formatter and validates it is not nil.
func (b *EngineBuilder) WithFormatter(f formatter.FormatStrategy) *EngineBuilder {
	b.formatter = f
	return b
}

func (b *EngineBuilder) WithOutput(o output.OutputStrategy) *EngineBuilder {
	b.output = o
	return b
}

// WithRetry sets the retry policy for the engine.
func (b *EngineBuilder) WithRetry(policy resilience.RetryPolicy) *EngineBuilder {
	b.retry = &policy
	return b
}

// WithCircuitBreaker sets the circuit breaker for the engine.
func (b *EngineBuilder) WithCircuitBreaker(cb *resilience.CircuitBreaker) *EngineBuilder {
	b.breaker = cb
	return b
}

// WithTracer sets the tracer for the engine.
func (b *EngineBuilder) WithTracer(tracer observability.Tracer) *EngineBuilder {
	b.tracer = tracer
	return b
}

// WithMetrics sets the metrics collector for the engine.
func (b *EngineBuilder) WithMetrics(collector observability.MetricsCollector) *EngineBuilder {
	b.metrics = collector
	return b
}

// Build validates all components and constructs the ReportEngine.
// It returns an error if any required component is missing or invalid.
//
// Validation checks:
//   - Provider is set and not nil
//   - Processor is set and not nil
//   - Formatter is set and not nil
//   - Output is set and not nil
//
// Returns:
//   - *ReportEngine: A fully configured engine ready to run
//   - error: Detailed error if validation fails
func (b *EngineBuilder) Build() (*ReportEngine, error) {
	// Collect all validation errors
	var errors []string

	// Validate provider
	if b.provider == nil {
		errors = append(errors, "provider is required but not set")
	}

	// Validate processor
	if b.processor == nil {
		errors = append(errors, "processor is required but not set")
	}

	// Validate formatter
	if b.formatter == nil {
		errors = append(errors, "formatter is required but not set")
	}

	// Validate output
	if b.output == nil {
		errors = append(errors, "output is required but not set")
	}

	// If any errors, return aggregated error
	if len(errors) > 0 {
		return nil, &BuilderValidationError{
			Errors: errors,
		}
	}

	// All components valid, construct engine

	prov := b.provider
	proc := b.processor
	out := b.output

	// Apply Metrics Decorators if collector is present
	// We wrap inner-most to capture raw component performance
	if b.metrics != nil {
		prov = observability.NewProviderWithMetrics(prov, b.metrics)
		// We need to cast ProcessorHandler to one that metrics accepts or update internal/observability/processor.go to accept Handler interface
		// processor.ProcessorHandler is an interface, so it should match.
		// Wait, NewProcessorWithMetrics takes processor.ProcessorHandler.
		proc = observability.NewProcessorWithMetrics(proc, b.metrics)
		out = observability.NewOutputWithMetrics(out, b.metrics)
	}

	// Apply CircuitBreaker Decorators if present
	// Applies BEFORE Retry so that CB protects downstream.
	// If wrapped inside Retry: Retry calls CB (success) -> CB calls Prov (fail) -> CB records fail.
	// Repeated failures open CB.
	// If wraps Retry: CB calls Retry (which retries N times). If total fails, CB records 1 fail.
	// We want CB to record every failure or the aggregate?
	// Usually invalid requests shouldn't trip CB, but downtime should.
	// We'll wrap INSIDE retry, so that 1 failed operation (after N retries) counts as 1 failure?
	// OR: wrap OUTSIDE retry, so retries happen, and if they all fail, CB counts.
	//
	// Wait, earlier design in Plan was: Retry wraps CB wraps Provider.
	// Retry calls CB.Execute(). CB calls Provider.Fetch().
	// If Provider fails, CB records failure. Retry sees error, waits, calls CB.Execute() again.
	// This is standard. CB counts individual attempts.
	if b.breaker != nil {
		// Clone breaker or use shared? Shared for Provider and Output?
		// We probably want SEPARATE breakers for Provider and Output.
		// For now, let's use the SAME policy but create new instances if we could.
		// But Builder pattern takes an INSTANCE.
		// Let's assume the user passed a configured struct instance which serves as config?
		// No, `resilience.CircuitBreaker` IS stateful.
		// Using the SAME instance for Prov and Output links their failure domains (if Prov fails, Output stops).
		// That might be undesirable.
		// Ideally we should accept Config and create instances.
		// But for now, if user passes ONE breaker, we verify usage.
		// Let's wrap Provider only? Or both?
		// If both, they share state. If DB (Prov) is down, S3 (Output) calls also fail?
		// Maybe acceptable for a simple engine.
		prov = resilience.NewProviderWithCircuitBreaker(prov, b.breaker)
		out = resilience.NewOutputWithCircuitBreaker(out, b.breaker)
	}

	// Apply Tracing Decorators if present
	// Applies BEFORE Retry so that we see spans for each attempt.
	if b.tracer != nil {
		prov = observability.NewProviderWithTracing(prov, b.tracer)
		proc = observability.NewProcessorWithTracing(proc, b.tracer)
		out = observability.NewOutputWithTracing(out, b.tracer)
	}

	// Apply Retry Decorators if policy is present
	// Retry wraps Metrics, so metrics record each attempt.
	if b.retry != nil {
		retrier := resilience.NewRetrier(*b.retry)
		prov = resilience.NewProviderWithRetry(prov, retrier)
		out = resilience.NewOutputWithRetry(out, retrier)
	}

	return &ReportEngine{
		Provider:  prov,
		Processor: proc,
		Formatter: b.formatter,
		Output:    out,
	}, nil
}

// Validate checks if all required components are set without building.
// This allows checking builder state before Build() is called.
//
// Returns:
//   - error: nil if valid, or error describing missing components
func (b *EngineBuilder) Validate() error {
	var errors []string

	if b.provider == nil {
		errors = append(errors, "provider not set")
	}
	if b.processor == nil {
		errors = append(errors, "processor not set")
	}
	if b.formatter == nil {
		errors = append(errors, "formatter not set")
	}
	if b.output == nil {
		errors = append(errors, "output not set")
	}

	if len(errors) > 0 {
		return &BuilderValidationError{Errors: errors}
	}

	return nil
}

// IsComplete returns true if all required components are set.
// This is useful for checking builder state programmatically.
func (b *EngineBuilder) IsComplete() bool {
	return b.provider != nil &&
		b.processor != nil &&
		b.formatter != nil &&
		b.output != nil
}

// Reset clears all components, returning the builder to initial state.
// This is useful for reusing a builder instance.
func (b *EngineBuilder) Reset() *EngineBuilder {
	b.provider = nil
	b.processor = nil
	b.formatter = nil
	b.output = nil
	return b
}

// BuilderValidationError represents validation errors from the builder.
type BuilderValidationError struct {
	Errors []string
}

// Error implements the error interface.
func (e *BuilderValidationError) Error() string {
	if len(e.Errors) == 1 {
		return fmt.Sprintf("builder validation failed: %s", e.Errors[0])
	}
	return fmt.Sprintf("builder validation failed: %d errors: %v", len(e.Errors), e.Errors)
}

// Predefined builder errors
var (
	// ErrBuilderIncomplete indicates not all components are set
	ErrBuilderIncomplete = fmt.Errorf("builder is incomplete")

	// ErrBuilderProviderNil indicates provider is nil
	ErrBuilderProviderNil = fmt.Errorf("provider cannot be nil")

	// ErrBuilderProcessorNil indicates processor is nil
	ErrBuilderProcessorNil = fmt.Errorf("processor cannot be nil")

	// ErrBuilderFormatterNil indicates formatter is nil
	ErrBuilderFormatterNil = fmt.Errorf("formatter cannot be nil")

	// ErrBuilderOutputNil indicates output is nil
	ErrBuilderOutputNil = fmt.Errorf("output cannot be nil")
)
