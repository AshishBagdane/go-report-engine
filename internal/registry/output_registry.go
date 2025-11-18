package registry

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/output"
)

type OutputFactory func() output.OutputStrategy

var outputRegistry = make(map[string]OutputFactory)

func RegisterOutput(name string, factory OutputFactory) {
	outputRegistry[name] = factory
}

func GetOutput(name string) (output.OutputStrategy, error) {
	if factory, ok := outputRegistry[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("output not found: %s", name)
}
