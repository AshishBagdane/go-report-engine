package registry

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/formatter"
)

type FormatterFactory func() formatter.FormatStrategy

var formatterRegistry = make(map[string]FormatterFactory)

func RegisterFormatter(name string, factory FormatterFactory) {
	formatterRegistry[name] = factory
}

func GetFormatter(name string) (formatter.FormatStrategy, error) {
	if factory, ok := formatterRegistry[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("formatter not found: %s", name)
}
