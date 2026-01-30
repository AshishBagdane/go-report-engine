package factory

import (
	"fmt"

	"github.com/AshishBagdane/go-report-engine/internal/engine"
	"github.com/AshishBagdane/go-report-engine/internal/registry"
)

// NewEngineFromConfig acts as the central Factory defined in your diagram.
// It reads the Config struct and uses the EngineBuilder to construct the engine.
func NewEngineFromConfig(cfg engine.Config) (*engine.ReportEngine, error) {
	// 1. Validate Config for required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 2. Create Components using Registries (Factory calls)

	// Provider
	prov, err := registry.GetProvider(cfg.Provider.Type) //
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	// Formatter
	fmtStrategy, err := registry.GetFormatter(cfg.Formatter.Type) //
	if err != nil {
		return nil, fmt.Errorf("formatter error: %w", err)
	}

	// Output
	outStrategy, err := registry.GetOutput(cfg.Output.Type) //
	if err != nil {
		return nil, fmt.Errorf("output error: %w", err)
	}

	// Processor Chain (Dynamic Creation using the processor_chain_factory)
	procChain, err := BuildProcessorChain(cfg.Processors)
	if err != nil {
		return nil, err
	}

	// 3. Assemble using the Builder
	return engine.NewEngineBuilder().
		WithProvider(prov).
		WithFormatter(fmtStrategy).
		WithOutput(outStrategy).
		WithProcessor(procChain).
		Build()
}
