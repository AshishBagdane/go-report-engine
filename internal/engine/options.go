package engine

import (
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// Option is a functional option for configuring the ReportEngine.
type Option func(*ReportEngine)

// WithProvider overrides the provider implementation.
func WithProvider(p provider.ProviderStrategy) Option {
	return func(r *ReportEngine) {
		r.Provider = p
	}
}

// WithProcessor overrides the processor pipeline.
func WithProcessor(p processor.ProcessorHandler) Option {
	return func(r *ReportEngine) {
		r.Processor = p
	}
}

// WithFormatter overrides the formatter implementation.
func WithFormatter(f formatter.FormatStrategy) Option {
	return func(r *ReportEngine) {
		r.Formatter = f
	}
}

// WithOutput overrides the output implementation.
func WithOutput(o output.OutputStrategy) Option {
	return func(r *ReportEngine) {
		r.Output = o
	}
}
