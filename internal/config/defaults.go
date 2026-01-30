// Package config provides default configuration presets for common use cases.
// These defaults make it easy to get started with sensible production-ready values.
package config

import "github.com/AshishBagdane/go-report-engine/internal/engine"

// DefaultConfig returns a minimal valid configuration with sensible defaults.
// This is the simplest configuration that will work out of the box.
//
// Default configuration:
//   - Provider: mock (with empty data)
//   - Processors: none (empty pipeline)
//   - Formatter: json (with 2-space indentation)
//   - Output: console
//
// Example:
//
//	config := config.DefaultConfig()
//	engine, err := factory.NewEngineFromConfig(config)
func DefaultConfig() engine.Config {
	return engine.Config{
		Provider: engine.ProviderConfig{
			Type:   "mock",
			Params: map[string]string{},
		},
		Processors: []engine.ProcessorConfig{},
		Formatter: engine.FormatterConfig{
			Type: "json",
			Params: map[string]string{
				"indent": "2",
			},
		},
		Output: engine.OutputConfig{
			Type:   "console",
			Params: map[string]string{},
		},
	}
}

// DefaultProviderConfig returns a default provider configuration.
// Uses mock provider with no parameters.
func DefaultProviderConfig() engine.ProviderConfig {
	return engine.ProviderConfig{
		Type:   "mock",
		Params: map[string]string{},
	}
}

// DefaultFormatterConfig returns a default formatter configuration.
// Uses JSON formatter with 2-space indentation.
func DefaultFormatterConfig() engine.FormatterConfig {
	return engine.FormatterConfig{
		Type: "json",
		Params: map[string]string{
			"indent": "2",
		},
	}
}

// DefaultOutputConfig returns a default output configuration.
// Uses console output with no parameters.
func DefaultOutputConfig() engine.OutputConfig {
	return engine.OutputConfig{
		Type:   "console",
		Params: map[string]string{},
	}
}

// ProductionConfig returns a production-ready configuration with best practices.
// This configuration is suitable for production use with proper error handling,
// validation, and structured output.
//
// Production configuration includes:
//   - Provider: Should be overridden with actual data source
//   - Processors: Includes validator and sanitizer
//   - Formatter: JSON with pretty printing disabled for performance
//   - Output: File output with proper error handling
//
// Example:
//
//	config := config.ProductionConfig()
//	config.Provider = engine.ProviderConfig{
//	    Type: "postgres",
//	    Params: map[string]string{
//	        "host": "db.example.com",
//	        "database": "reports",
//	    },
//	}
func ProductionConfig() engine.Config {
	return engine.Config{
		Provider: engine.ProviderConfig{
			Type:   "mock",
			Params: map[string]string{},
		},
		Processors: []engine.ProcessorConfig{
			{
				Type: "validator",
				Params: map[string]string{
					"strict": "true",
				},
			},
		},
		Formatter: engine.FormatterConfig{
			Type: "json",
			Params: map[string]string{
				"indent": "", // No indentation for production
			},
		},
		Output: engine.OutputConfig{
			Type: "file",
			Params: map[string]string{
				"path": "/var/log/reports/output.json",
				"mode": "0644",
			},
		},
	}
}

// DevelopmentConfig returns a development-friendly configuration.
// This configuration is optimized for debugging and local development
// with verbose output and human-readable formatting.
//
// Development configuration includes:
//   - Provider: Mock provider for quick testing
//   - Processors: None (for simplicity)
//   - Formatter: JSON with pretty printing
//   - Output: Console for immediate feedback
//
// Example:
//
//	config := config.DevelopmentConfig()
//	engine, err := factory.NewEngineFromConfig(config)
func DevelopmentConfig() engine.Config {
	return engine.Config{
		Provider: engine.ProviderConfig{
			Type: "mock",
			Params: map[string]string{
				"verbose": "true",
			},
		},
		Processors: []engine.ProcessorConfig{},
		Formatter: engine.FormatterConfig{
			Type: "json",
			Params: map[string]string{
				"indent": "4", // Extra indentation for readability
				"pretty": "true",
			},
		},
		Output: engine.OutputConfig{
			Type:   "console",
			Params: map[string]string{},
		},
	}
}

// TestingConfig returns a configuration optimized for automated testing.
// This configuration is minimal and predictable for unit and integration tests.
//
// Testing configuration includes:
//   - Provider: Mock with empty data
//   - Processors: None
//   - Formatter: JSON without indentation
//   - Output: Console
//
// Example:
//
//	config := config.TestingConfig()
//	engine, err := factory.NewEngineFromConfig(config)
func TestingConfig() engine.Config {
	return engine.Config{
		Provider: engine.ProviderConfig{
			Type:   "mock",
			Params: map[string]string{},
		},
		Processors: []engine.ProcessorConfig{},
		Formatter: engine.FormatterConfig{
			Type:   "json",
			Params: map[string]string{},
		},
		Output: engine.OutputConfig{
			Type:   "console",
			Params: map[string]string{},
		},
	}
}

// CSVConfig returns a configuration for CSV output.
// This is a common use case for data exports and reports.
//
// CSV configuration includes:
//   - Provider: Should be overridden with actual data source
//   - Processors: None (customize as needed)
//   - Formatter: CSV with headers
//   - Output: File
//
// Example:
//
//	config := config.CSVConfig()
//	config.Provider.Type = "postgres"
//	config.Output.Params["path"] = "/tmp/report.csv"
func CSVConfig() engine.Config {
	return engine.Config{
		Provider: engine.ProviderConfig{
			Type:   "mock",
			Params: map[string]string{},
		},
		Processors: []engine.ProcessorConfig{},
		Formatter: engine.FormatterConfig{
			Type: "csv",
			Params: map[string]string{
				"delimiter":      ",",
				"include_header": "true",
			},
		},
		Output: engine.OutputConfig{
			Type: "file",
			Params: map[string]string{
				"path": "output.csv",
				"mode": "0644",
			},
		},
	}
}

// ConfigWithProcessor returns a new config with an additional processor.
// This is a helper function for building configs with processing pipelines.
//
// Parameters:
//   - base: The base configuration to extend
//   - processorType: The type of processor to add
//   - params: Parameters for the processor
//
// Returns a new Config with the processor appended to the pipeline.
//
// Example:
//
//	config := config.DefaultConfig()
//	config = config.ConfigWithProcessor(config, "filter", map[string]string{
//	    "min_score": "90",
//	})
func ConfigWithProcessor(base engine.Config, processorType string, params map[string]string) engine.Config {
	// Make a copy to avoid modifying the original
	newConfig := base

	// Copy processors slice
	processors := make([]engine.ProcessorConfig, len(base.Processors))
	copy(processors, base.Processors)

	// Append new processor
	processors = append(processors, engine.ProcessorConfig{
		Type:   processorType,
		Params: params,
	})

	newConfig.Processors = processors
	return newConfig
}

// ConfigWithProviderParams returns a new config with provider params merged.
// This is useful for adding or updating provider parameters without replacing
// the entire provider configuration.
//
// Parameters:
//   - base: The base configuration to extend
//   - params: Parameters to merge into provider params
//
// Returns a new Config with merged provider parameters.
//
// Example:
//
//	config := config.DefaultConfig()
//	config = config.ConfigWithProviderParams(config, map[string]string{
//	    "host": "localhost",
//	    "port": "5432",
//	})
func ConfigWithProviderParams(base engine.Config, params map[string]string) engine.Config {
	// Make a copy
	newConfig := base

	// Initialize params if nil
	if newConfig.Provider.Params == nil {
		newConfig.Provider.Params = make(map[string]string)
	}

	// Copy existing params
	mergedParams := make(map[string]string, len(base.Provider.Params)+len(params))
	for k, v := range base.Provider.Params {
		mergedParams[k] = v
	}

	// Merge new params
	for k, v := range params {
		mergedParams[k] = v
	}

	newConfig.Provider.Params = mergedParams
	return newConfig
}

// ConfigWithFormatterParams returns a new config with formatter params merged.
// Similar to ConfigWithProviderParams but for formatter configuration.
//
// Example:
//
//	config := config.DefaultConfig()
//	config = config.ConfigWithFormatterParams(config, map[string]string{
//	    "indent": "4",
//	    "pretty": "true",
//	})
func ConfigWithFormatterParams(base engine.Config, params map[string]string) engine.Config {
	newConfig := base

	if newConfig.Formatter.Params == nil {
		newConfig.Formatter.Params = make(map[string]string)
	}

	mergedParams := make(map[string]string, len(base.Formatter.Params)+len(params))
	for k, v := range base.Formatter.Params {
		mergedParams[k] = v
	}

	for k, v := range params {
		mergedParams[k] = v
	}

	newConfig.Formatter.Params = mergedParams
	return newConfig
}

// ConfigWithOutputParams returns a new config with output params merged.
// Similar to ConfigWithProviderParams but for output configuration.
//
// Example:
//
//	config := config.DefaultConfig()
//	config = config.ConfigWithOutputParams(config, map[string]string{
//	    "path": "/tmp/output.json",
//	})
func ConfigWithOutputParams(base engine.Config, params map[string]string) engine.Config {
	newConfig := base

	if newConfig.Output.Params == nil {
		newConfig.Output.Params = make(map[string]string)
	}

	mergedParams := make(map[string]string, len(base.Output.Params)+len(params))
	for k, v := range base.Output.Params {
		mergedParams[k] = v
	}

	for k, v := range params {
		mergedParams[k] = v
	}

	newConfig.Output.Params = mergedParams
	return newConfig
}
