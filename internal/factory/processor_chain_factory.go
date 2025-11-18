package factory

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/engine"    // For ProcessorConfig
	"github.com/AshishBagdane/report-engine/internal/processor" // For ProcessorHandler
	"github.com/AshishBagdane/report-engine/internal/registry"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// BuildProcessorChain reads a list of configurations and links them together
// using the Chain of Responsibility pattern.
func BuildProcessorChain(configs []engine.ProcessorConfig) (processor.ProcessorHandler, error) {
	if len(configs) == 0 {
		// Return a default base processor if no chain is defined
		return &processor.BaseProcessor{}, nil //
	}

	var head, current processor.ProcessorHandler

	for i, cfg := range configs {
		// 1. Get the factory instance from the registry (this returns a wrapper like FilterWrapper)
		procInstance, err := registry.GetProcessor(cfg.Type)
		if err != nil {
			return nil, fmt.Errorf("step %d ('%s') factory failed: %w", i, cfg.Type, err)
		}

		// 2. Configure the instance if it's configurable
		// The wrappers implement Configure to pass params to the user's strategy
		if configurable, ok := procInstance.(api.Configurable); ok {
			if err := configurable.Configure(cfg.Params); err != nil {
				return nil, fmt.Errorf("step %d ('%s') configuration failed: %w", i, cfg.Type, err)
			}
		}

		// 3. Link the chain
		if head == nil {
			head = procInstance
			current = procInstance
		} else {
			current.SetNext(procInstance) //
			current = procInstance
		}
	}

	return head, nil
}
