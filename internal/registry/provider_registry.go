package registry

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/provider"
)

type ProviderFactory func() provider.ProviderStrategy

var providerRegistry = make(map[string]ProviderFactory)

func RegisterProvider(name string, factory ProviderFactory) {
	providerRegistry[name] = factory
}

func GetProvider(name string) (provider.ProviderStrategy, error) {
	if factory, ok := providerRegistry[name]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}
