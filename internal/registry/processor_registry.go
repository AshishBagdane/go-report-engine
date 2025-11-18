package registry

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// The original generic factory and registry
type ProcessorFactory func() processor.ProcessorHandler

var processorRegistry = make(map[string]ProcessorFactory)

func RegisterProcessor(name string, factory ProcessorFactory) {
	processorRegistry[name] = factory
}

func GetProcessor(name string) (processor.ProcessorHandler, error) {
	if factory, ok := processorRegistry[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("processor not found: %s", name)
}

// --- NEW: Type-Safe Registration Helpers ---

// RegisterFilter allows users to register a simple FilterStrategy,
// automatically wrapping it in a FilterWrapper.
func RegisterFilter(name string, s api.FilterStrategy) {
	RegisterProcessor(name, func() processor.ProcessorHandler {
		return processor.NewFilterWrapper(s)
	})
}

// RegisterValidator allows users to register a simple ValidatorStrategy,
// automatically wrapping it in a ValidatorWrapper.
func RegisterValidator(name string, s api.ValidatorStrategy) {
	RegisterProcessor(name, func() processor.ProcessorHandler {
		return processor.NewValidatorWrapper(s)
	})
}

// RegisterTransformer allows users to register a simple TransformerStrategy,
// automatically wrapping it in a TransformWrapper.
func RegisterTransformer(name string, s api.TransformerStrategy) {
	RegisterProcessor(name, func() processor.ProcessorHandler {
		return processor.NewTransformWrapper(s)
	})
}
