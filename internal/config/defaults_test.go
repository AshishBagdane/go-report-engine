package config

import (
	"testing"

	"github.com/AshishBagdane/report-engine/internal/engine"
)

// TestDefaultConfig tests default configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Verify structure
	if config.Provider.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock", config.Provider.Type)
	}

	if config.Formatter.Type != "json" {
		t.Errorf("Formatter type = %s, expected json", config.Formatter.Type)
	}

	if config.Output.Type != "console" {
		t.Errorf("Output type = %s, expected console", config.Output.Type)
	}

	// Verify it's valid
	if err := config.Validate(); err != nil {
		t.Errorf("DefaultConfig() should be valid, got: %v", err)
	}

	// Verify JSON formatter has indent
	if config.Formatter.Params["indent"] != "2" {
		t.Error("Default JSON formatter should have indent=2")
	}
}

// TestDefaultProviderConfig tests default provider config
func TestDefaultProviderConfig(t *testing.T) {
	config := DefaultProviderConfig()

	if config.Type != "mock" {
		t.Errorf("Provider type = %s, expected mock", config.Type)
	}

	if config.Params == nil {
		t.Error("Provider params should not be nil")
	}
}

// TestDefaultFormatterConfig tests default formatter config
func TestDefaultFormatterConfig(t *testing.T) {
	config := DefaultFormatterConfig()

	if config.Type != "json" {
		t.Errorf("Formatter type = %s, expected json", config.Type)
	}

	if config.Params["indent"] != "2" {
		t.Error("Default formatter should have indent=2")
	}
}

// TestDefaultOutputConfig tests default output config
func TestDefaultOutputConfig(t *testing.T) {
	config := DefaultOutputConfig()

	if config.Type != "console" {
		t.Errorf("Output type = %s, expected console", config.Type)
	}

	if config.Params == nil {
		t.Error("Output params should not be nil")
	}
}

// TestProductionConfig tests production configuration
func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	// Verify structure
	if config.Provider.Type == "" {
		t.Error("Production config should have provider type")
	}

	if len(config.Processors) == 0 {
		t.Error("Production config should have processors")
	}

	if config.Formatter.Type != "json" {
		t.Error("Production config should use JSON formatter")
	}

	if config.Output.Type != "file" {
		t.Error("Production config should use file output")
	}

	// Verify it's valid
	if err := config.Validate(); err != nil {
		t.Errorf("ProductionConfig() should be valid, got: %v", err)
	}

	// Verify JSON has no indentation for performance
	if config.Formatter.Params["indent"] != "" {
		t.Error("Production JSON should have no indentation")
	}

	// Verify validator is included
	found := false
	for _, proc := range config.Processors {
		if proc.Type == "validator" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Production config should include validator processor")
	}
}

// TestDevelopmentConfig tests development configuration
func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	// Verify structure
	if config.Provider.Type != "mock" {
		t.Error("Development config should use mock provider")
	}

	if config.Formatter.Type != "json" {
		t.Error("Development config should use JSON formatter")
	}

	if config.Output.Type != "console" {
		t.Error("Development config should use console output")
	}

	// Verify it's valid
	if err := config.Validate(); err != nil {
		t.Errorf("DevelopmentConfig() should be valid, got: %v", err)
	}

	// Verify JSON has extra indentation for readability
	if config.Formatter.Params["indent"] != "4" {
		t.Error("Development JSON should have indent=4")
	}

	if config.Formatter.Params["pretty"] != "true" {
		t.Error("Development JSON should have pretty=true")
	}
}

// TestTestingConfig tests testing configuration
func TestTestingConfig(t *testing.T) {
	config := TestingConfig()

	// Verify structure
	if config.Provider.Type != "mock" {
		t.Error("Testing config should use mock provider")
	}

	if len(config.Processors) != 0 {
		t.Error("Testing config should have no processors")
	}

	if config.Formatter.Type != "json" {
		t.Error("Testing config should use JSON formatter")
	}

	if config.Output.Type != "console" {
		t.Error("Testing config should use console output")
	}

	// Verify it's valid
	if err := config.Validate(); err != nil {
		t.Errorf("TestingConfig() should be valid, got: %v", err)
	}
}

// TestCSVConfig tests CSV configuration
func TestCSVConfig(t *testing.T) {
	config := CSVConfig()

	// Verify structure
	if config.Formatter.Type != "csv" {
		t.Errorf("CSV config formatter type = %s, expected csv", config.Formatter.Type)
	}

	if config.Output.Type != "file" {
		t.Error("CSV config should use file output")
	}

	// Verify it's valid
	if err := config.Validate(); err != nil {
		t.Errorf("CSVConfig() should be valid, got: %v", err)
	}

	// Verify CSV params
	if config.Formatter.Params["delimiter"] != "," {
		t.Error("CSV config should have comma delimiter")
	}

	if config.Formatter.Params["include_header"] != "true" {
		t.Error("CSV config should include headers")
	}
}

// TestConfigWithProcessor tests adding processors
func TestConfigWithProcessor(t *testing.T) {
	base := DefaultConfig()

	config := ConfigWithProcessor(base, "filter", map[string]string{
		"min_score": "90",
	})

	// Original should be unchanged
	if len(base.Processors) != 0 {
		t.Error("Original config should be unchanged")
	}

	// New config should have processor
	if len(config.Processors) != 1 {
		t.Fatalf("New config should have 1 processor, got %d", len(config.Processors))
	}

	proc := config.Processors[0]
	if proc.Type != "filter" {
		t.Errorf("Processor type = %s, expected filter", proc.Type)
	}

	if proc.Params["min_score"] != "90" {
		t.Error("Processor params not set correctly")
	}
}

// TestConfigWithProcessorMultiple tests adding multiple processors
func TestConfigWithProcessorMultiple(t *testing.T) {
	config := DefaultConfig()

	config = ConfigWithProcessor(config, "filter", map[string]string{"min": "80"})
	config = ConfigWithProcessor(config, "validator", map[string]string{"strict": "true"})
	config = ConfigWithProcessor(config, "transformer", map[string]string{"format": "upper"})

	if len(config.Processors) != 3 {
		t.Fatalf("Config should have 3 processors, got %d", len(config.Processors))
	}

	// Verify order
	if config.Processors[0].Type != "filter" {
		t.Error("First processor should be filter")
	}
	if config.Processors[1].Type != "validator" {
		t.Error("Second processor should be validator")
	}
	if config.Processors[2].Type != "transformer" {
		t.Error("Third processor should be transformer")
	}
}

// TestConfigWithProviderParams tests adding provider params
func TestConfigWithProviderParams(t *testing.T) {
	base := DefaultConfig()

	config := ConfigWithProviderParams(base, map[string]string{
		"host": "localhost",
		"port": "5432",
	})

	// Original should be unchanged
	if len(base.Provider.Params) != 0 {
		t.Error("Original config should be unchanged")
	}

	// New config should have params
	if config.Provider.Params["host"] != "localhost" {
		t.Error("Provider param 'host' not set")
	}

	if config.Provider.Params["port"] != "5432" {
		t.Error("Provider param 'port' not set")
	}
}

// TestConfigWithProviderParamsMerge tests merging provider params
func TestConfigWithProviderParamsMerge(t *testing.T) {
	base := engine.Config{
		Provider: engine.ProviderConfig{
			Type: "postgres",
			Params: map[string]string{
				"host": "original",
				"port": "5432",
			},
		},
		Processors: []engine.ProcessorConfig{},
		Formatter:  DefaultFormatterConfig(),
		Output:     DefaultOutputConfig(),
	}

	config := ConfigWithProviderParams(base, map[string]string{
		"host":     "updated",
		"database": "mydb",
	})

	// Original params should be unchanged
	if base.Provider.Params["host"] != "original" {
		t.Error("Original config should be unchanged")
	}

	// New config should have merged params
	if config.Provider.Params["host"] != "updated" {
		t.Error("Provider param 'host' should be updated")
	}

	if config.Provider.Params["port"] != "5432" {
		t.Error("Provider param 'port' should be preserved")
	}

	if config.Provider.Params["database"] != "mydb" {
		t.Error("Provider param 'database' should be added")
	}
}

// TestConfigWithFormatterParams tests adding formatter params
func TestConfigWithFormatterParams(t *testing.T) {
	base := DefaultConfig()

	config := ConfigWithFormatterParams(base, map[string]string{
		"indent": "4",
		"pretty": "true",
	})

	// New config should have params
	if config.Formatter.Params["indent"] != "4" {
		t.Error("Formatter param 'indent' not set correctly")
	}

	if config.Formatter.Params["pretty"] != "true" {
		t.Error("Formatter param 'pretty' not set")
	}
}

// TestConfigWithFormatterParamsMerge tests merging formatter params
func TestConfigWithFormatterParamsMerge(t *testing.T) {
	base := engine.Config{
		Provider:   DefaultProviderConfig(),
		Processors: []engine.ProcessorConfig{},
		Formatter: engine.FormatterConfig{
			Type: "json",
			Params: map[string]string{
				"indent": "2",
			},
		},
		Output: DefaultOutputConfig(),
	}

	config := ConfigWithFormatterParams(base, map[string]string{
		"indent": "4",
		"pretty": "true",
	})

	// Original should be unchanged
	if base.Formatter.Params["indent"] != "2" {
		t.Error("Original config should be unchanged")
	}

	// New config should have merged params
	if config.Formatter.Params["indent"] != "4" {
		t.Error("Formatter param 'indent' should be updated")
	}

	if config.Formatter.Params["pretty"] != "true" {
		t.Error("Formatter param 'pretty' should be added")
	}
}

// TestConfigWithOutputParams tests adding output params
func TestConfigWithOutputParams(t *testing.T) {
	base := DefaultConfig()

	config := ConfigWithOutputParams(base, map[string]string{
		"path": "/tmp/output.json",
		"mode": "0644",
	})

	// New config should have params
	if config.Output.Params["path"] != "/tmp/output.json" {
		t.Error("Output param 'path' not set")
	}

	if config.Output.Params["mode"] != "0644" {
		t.Error("Output param 'mode' not set")
	}
}

// TestConfigWithOutputParamsMerge tests merging output params
func TestConfigWithOutputParamsMerge(t *testing.T) {
	base := engine.Config{
		Provider:   DefaultProviderConfig(),
		Processors: []engine.ProcessorConfig{},
		Formatter:  DefaultFormatterConfig(),
		Output: engine.OutputConfig{
			Type: "file",
			Params: map[string]string{
				"path": "/tmp/original.json",
			},
		},
	}

	config := ConfigWithOutputParams(base, map[string]string{
		"path": "/tmp/updated.json",
		"mode": "0644",
	})

	// Original should be unchanged
	if base.Output.Params["path"] != "/tmp/original.json" {
		t.Error("Original config should be unchanged")
	}

	// New config should have merged params
	if config.Output.Params["path"] != "/tmp/updated.json" {
		t.Error("Output param 'path' should be updated")
	}

	if config.Output.Params["mode"] != "0644" {
		t.Error("Output param 'mode' should be added")
	}
}

// TestAllDefaultsAreValid tests that all default configs are valid
func TestAllDefaultsAreValid(t *testing.T) {
	configs := map[string]engine.Config{
		"DefaultConfig":     DefaultConfig(),
		"ProductionConfig":  ProductionConfig(),
		"DevelopmentConfig": DevelopmentConfig(),
		"TestingConfig":     TestingConfig(),
		"CSVConfig":         CSVConfig(),
	}

	for name, config := range configs {
		t.Run(name, func(t *testing.T) {
			if err := config.Validate(); err != nil {
				t.Errorf("%s should be valid, got: %v", name, err)
			}
		})
	}
}

// TestConfigBuilderPattern tests building configs step by step
func TestConfigBuilderPattern(t *testing.T) {
	// Start with default
	config := DefaultConfig()

	// Add provider params
	config = ConfigWithProviderParams(config, map[string]string{
		"host": "localhost",
		"port": "5432",
	})

	// Add processors
	config = ConfigWithProcessor(config, "filter", map[string]string{
		"min_score": "90",
	})
	config = ConfigWithProcessor(config, "validator", map[string]string{
		"strict": "true",
	})

	// Update formatter params
	config = ConfigWithFormatterParams(config, map[string]string{
		"indent": "4",
	})

	// Update output params
	config = ConfigWithOutputParams(config, map[string]string{
		"path": "/tmp/report.json",
	})

	// Verify final config
	if len(config.Provider.Params) != 2 {
		t.Error("Provider should have 2 params")
	}

	if len(config.Processors) != 2 {
		t.Error("Config should have 2 processors")
	}

	if config.Formatter.Params["indent"] != "4" {
		t.Error("Formatter indent should be updated")
	}

	if config.Output.Params["path"] != "/tmp/report.json" {
		t.Error("Output path should be set")
	}

	// Should still be valid
	if err := config.Validate(); err != nil {
		t.Errorf("Built config should be valid, got: %v", err)
	}
}

// TestConfigImmutability tests that helper functions don't modify originals
func TestConfigImmutability(t *testing.T) {
	original := DefaultConfig()
	originalProviderParamsLen := len(original.Provider.Params)
	originalProcessorsLen := len(original.Processors)

	// Apply transformations
	_ = ConfigWithProviderParams(original, map[string]string{"key": "value"})
	_ = ConfigWithProcessor(original, "filter", map[string]string{})
	_ = ConfigWithFormatterParams(original, map[string]string{"key": "value"})
	_ = ConfigWithOutputParams(original, map[string]string{"key": "value"})

	// Original should be unchanged
	if len(original.Provider.Params) != originalProviderParamsLen {
		t.Error("Original provider params were modified")
	}

	if len(original.Processors) != originalProcessorsLen {
		t.Error("Original processors were modified")
	}
}

// BenchmarkDefaultConfig benchmarks default config creation
func BenchmarkDefaultConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultConfig()
	}
}

// BenchmarkProductionConfig benchmarks production config creation
func BenchmarkProductionConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProductionConfig()
	}
}

// BenchmarkConfigWithProcessor benchmarks adding processors
func BenchmarkConfigWithProcessor(b *testing.B) {
	base := DefaultConfig()
	params := map[string]string{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConfigWithProcessor(base, "filter", params)
	}
}

// BenchmarkConfigWithProviderParams benchmarks adding provider params
func BenchmarkConfigWithProviderParams(b *testing.B) {
	base := DefaultConfig()
	params := map[string]string{"host": "localhost", "port": "5432"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConfigWithProviderParams(base, params)
	}
}
