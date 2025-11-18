package engine

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

type EngineBuilder struct {
	provider  provider.ProviderStrategy
	processor processor.ProcessorHandler
	formatter formatter.FormatStrategy
	output    output.OutputStrategy
}

// NewEngineBuilder
func NewEngineBuilder() *EngineBuilder {
	return &EngineBuilder{}
}

// WithProvider
func (b *EngineBuilder) WithProvider(p provider.ProviderStrategy) *EngineBuilder {
	b.provider = p
	return b
}

// WithProcessor
func (b *EngineBuilder) WithProcessor(p processor.ProcessorHandler) *EngineBuilder {
	b.processor = p
	return b
}

// WithFormatter
func (b *EngineBuilder) WithFormatter(f formatter.FormatStrategy) *EngineBuilder {
	b.formatter = f
	return b
}

// WithOutput
func (b *EngineBuilder) WithOutput(o output.OutputStrategy) *EngineBuilder {
	b.output = o
	return b
}

// Build validates everything and returns a ready-to-use engine
func (b *EngineBuilder) Build() (*ReportEngine, error) {
	if b.provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	if b.processor == nil {
		return nil, fmt.Errorf("processor is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("formatter is required")
	}
	if b.output == nil {
		return nil, fmt.Errorf("output is required")
	}

	return &ReportEngine{
		Provider:  b.provider,
		Processor: b.processor,
		Formatter: b.formatter,
		Output:    b.output,
	}, nil
}
