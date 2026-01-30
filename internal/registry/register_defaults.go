package registry

import (
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// init registers the default providers.
// This allows them to be available without manual registration in basic usage.
func init() {
	// Register mock provider
	// Note: factories used in tests are often specific, but having a default one is good.
	// However, MockProvider usually needs data passed to constructor.
	// So we might register a factory that returns a clean mock?
	// provider.NewMockProvider([]map[string]interface{}{})

	// Register CSV provider
	RegisterProvider("csv", func() provider.ProviderStrategy {
		return provider.NewCSVProvider()
	})

	// Register Mock provider with empty data as default
	RegisterProvider("mock", func() provider.ProviderStrategy {
		return provider.NewMockProvider(nil)
	})

	// Register SQL/Database provider
	RegisterProvider("sql", func() provider.ProviderStrategy {
		return provider.NewSQLProvider()
	})
	// Alias 'database' for convenience
	RegisterProvider("database", func() provider.ProviderStrategy {
		return provider.NewSQLProvider()
	})

	// Register REST/API provider
	RegisterProvider("rest", func() provider.ProviderStrategy {
		return provider.NewRESTProvider()
	})
	RegisterProvider("api", func() provider.ProviderStrategy {
		return provider.NewRESTProvider()
	})
	RegisterProvider("http", func() provider.ProviderStrategy {
		return provider.NewRESTProvider()
	})

	// Register CSV Formatter
	RegisterFormatter("csv", func() formatter.FormatStrategy {
		return formatter.NewCSVFormatter()
	})

	// Register YAML Formatter
	RegisterFormatter("yaml", func() formatter.FormatStrategy {
		return formatter.NewYAMLFormatter()
	})
	RegisterFormatter("yml", func() formatter.FormatStrategy {
		return formatter.NewYAMLFormatter()
	})

	// Register File Output
	RegisterOutput("file", func() output.OutputStrategy {
		return output.NewFileOutput()
	})

	// Register Processors
	RegisterProcessor("deduplicate", func() processor.ProcessorHandler {
		return processor.NewDeduplicateProcessor(nil)
	})
	RegisterProcessor("aggregate", func() processor.ProcessorHandler {
		return processor.NewAggregateProcessor(nil, nil)
	})
}
