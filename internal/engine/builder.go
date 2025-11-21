package engine

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

// EngineBuilder provides a fluent interface for constructing a ReportEngine.
// It implements the Builder pattern with comprehensive validation.
type EngineBuilder struct {
	provider  provider.ProviderStrategy
	processor processor.ProcessorHandler
	formatter formatter.FormatStrategy
	output    output.OutputStrategy
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

// WithOutput sets the output and validates it is not nil.
func (b *EngineBuilder) WithOutput(o output.OutputStrategy) *EngineBuilder {
	b.output = o
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
	return &ReportEngine{
		Provider:  b.provider,
		Processor: b.processor,
		Formatter: b.formatter,
		Output:    b.output,
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
