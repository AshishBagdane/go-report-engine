// Package config provides integration helpers that connect configuration loading
// with the engine factory for seamless engine creation from config files.
package config

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"github.com/AshishBagdane/report-engine/internal/factory"
)

// LoadAndBuild loads a configuration file and builds an engine in one step.
// This is a convenience function that combines LoadFromFile and NewEngineFromConfig.
//
// Parameters:
//   - path: Path to the configuration file (YAML or JSON)
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//   - error: nil on success, error describing the failure otherwise
//
// Example:
//
//	engine, err := config.LoadAndBuild("config.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to create engine: %v", err)
//	}
//	engine.Run()
func LoadAndBuild(path string) (*engine.ReportEngine, error) {
	// Load configuration from file
	cfg, err := LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Build engine from config
	eng, err := factory.NewEngineFromConfig(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build engine: %w", err)
	}

	return eng, nil
}

// LoadAndBuildWithEnv loads a configuration file with environment overrides
// and builds an engine in one step.
//
// This function combines LoadFromFileWithEnv and NewEngineFromConfig.
// Environment variables will override values from the configuration file.
//
// Parameters:
//   - path: Path to the configuration file (YAML or JSON)
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//   - error: nil on success, error describing the failure otherwise
//
// Example:
//
//	// Set environment variables
//	os.Setenv("ENGINE_PROVIDER_TYPE", "postgres")
//	os.Setenv("ENGINE_PROVIDER_PARAM_HOST", "localhost")
//
//	engine, err := config.LoadAndBuildWithEnv("config.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to create engine: %v", err)
//	}
//	engine.Run()
func LoadAndBuildWithEnv(path string) (*engine.ReportEngine, error) {
	// Load configuration with environment overrides
	cfg, err := LoadFromFileWithEnv(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Build engine from config
	eng, err := factory.NewEngineFromConfig(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build engine: %w", err)
	}

	return eng, nil
}

// BuildFromBytes loads configuration from raw bytes and builds an engine.
// This is useful when configuration is embedded or retrieved from external sources.
//
// Parameters:
//   - data: Raw configuration data
//   - format: "yaml" or "json"
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//   - error: nil on success, error describing the failure otherwise
//
// Example:
//
//	yamlConfig := []byte(`
//	  provider: {type: mock}
//	  formatter: {type: json}
//	  output: {type: console}
//	`)
//	engine, err := config.BuildFromBytes(yamlConfig, "yaml")
func BuildFromBytes(data []byte, format string) (*engine.ReportEngine, error) {
	// Load configuration from bytes
	cfg, err := LoadFromBytes(data, format)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Build engine from config
	eng, err := factory.NewEngineFromConfig(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build engine: %w", err)
	}

	return eng, nil
}

// BuildFromDefault creates an engine using default configuration.
// This is the quickest way to get a working engine for testing or simple use cases.
//
// Returns:
//   - *engine.ReportEngine: Engine with default configuration
//   - error: nil on success, error if default config is invalid (should never happen)
//
// Example:
//
//	engine, err := config.BuildFromDefault()
//	if err != nil {
//	    log.Fatalf("Failed to create engine: %v", err)
//	}
//	engine.Run()
func BuildFromDefault() (*engine.ReportEngine, error) {
	cfg := DefaultConfig()
	return factory.NewEngineFromConfig(cfg)
}

// BuildFromProduction creates an engine using production configuration.
// The provider should be customized for your specific data source.
//
// Returns:
//   - *engine.ReportEngine: Engine with production configuration
//   - error: nil on success, error if config is invalid
//
// Example:
//
//	engine, err := config.BuildFromProduction()
//	if err != nil {
//	    log.Fatalf("Failed to create engine: %v", err)
//	}
//	engine.Run()
func BuildFromProduction() (*engine.ReportEngine, error) {
	cfg := ProductionConfig()
	return factory.NewEngineFromConfig(cfg)
}

// BuildFromDevelopment creates an engine using development configuration.
// This is optimized for local development with verbose output.
//
// Returns:
//   - *engine.ReportEngine: Engine with development configuration
//   - error: nil on success, error if config is invalid
//
// Example:
//
//	engine, err := config.BuildFromDevelopment()
//	if err != nil {
//	    log.Fatalf("Failed to create engine: %v", err)
//	}
//	engine.Run()
func BuildFromDevelopment() (*engine.ReportEngine, error) {
	cfg := DevelopmentConfig()
	return factory.NewEngineFromConfig(cfg)
}

// BuildFromTesting creates an engine using testing configuration.
// This is optimized for unit and integration tests.
//
// Returns:
//   - *engine.ReportEngine: Engine with testing configuration
//   - error: nil on success, error if config is invalid
//
// Example:
//
//	engine, err := config.BuildFromTesting()
//	if err != nil {
//	    t.Fatalf("Failed to create engine: %v", err)
//	}
//	engine.Run()
func BuildFromTesting() (*engine.ReportEngine, error) {
	cfg := TestingConfig()
	return factory.NewEngineFromConfig(cfg)
}

// MustLoadAndBuild is like LoadAndBuild but panics on error.
// This is useful for initialization code where failure should stop the program.
//
// Parameters:
//   - path: Path to the configuration file
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//
// Panics if the config cannot be loaded or engine cannot be built.
//
// Example:
//
//	func init() {
//	    engine := config.MustLoadAndBuild("config.yaml")
//	    // Use engine...
//	}
func MustLoadAndBuild(path string) *engine.ReportEngine {
	eng, err := LoadAndBuild(path)
	if err != nil {
		panic(fmt.Sprintf("MustLoadAndBuild failed: %v", err))
	}
	return eng
}

// MustLoadAndBuildWithEnv is like LoadAndBuildWithEnv but panics on error.
// This is useful for initialization code where failure should stop the program.
//
// Parameters:
//   - path: Path to the configuration file
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//
// Panics if the config cannot be loaded or engine cannot be built.
//
// Example:
//
//	func init() {
//	    engine := config.MustLoadAndBuildWithEnv("config.yaml")
//	    // Use engine...
//	}
func MustLoadAndBuildWithEnv(path string) *engine.ReportEngine {
	eng, err := LoadAndBuildWithEnv(path)
	if err != nil {
		panic(fmt.Sprintf("MustLoadAndBuildWithEnv failed: %v", err))
	}
	return eng
}

// MustBuildFromDefault is like BuildFromDefault but panics on error.
// This is useful for initialization code where failure should stop the program.
//
// Returns:
//   - *engine.ReportEngine: Engine with default configuration
//
// Panics if the engine cannot be built (should never happen with valid defaults).
//
// Example:
//
//	func init() {
//	    engine := config.MustBuildFromDefault()
//	    // Use engine...
//	}
func MustBuildFromDefault() *engine.ReportEngine {
	eng, err := BuildFromDefault()
	if err != nil {
		panic(fmt.Sprintf("MustBuildFromDefault failed: %v", err))
	}
	return eng
}

// MustBuildFromProduction is like BuildFromProduction but panics on error.
// This is useful for initialization code where failure should stop the program.
//
// Returns:
//   - *engine.ReportEngine: Engine with production configuration
//
// Panics if the engine cannot be built.
//
// Example:
//
//	func init() {
//	    engine := config.MustBuildFromProduction()
//	    // Use engine...
//	}
func MustBuildFromProduction() *engine.ReportEngine {
	eng, err := BuildFromProduction()
	if err != nil {
		panic(fmt.Sprintf("MustBuildFromProduction failed: %v", err))
	}
	return eng
}

// MustBuildFromDevelopment is like BuildFromDevelopment but panics on error.
// This is useful for initialization code where failure should stop the program.
//
// Returns:
//   - *engine.ReportEngine: Engine with development configuration
//
// Panics if the engine cannot be built.
//
// Example:
//
//	func init() {
//	    engine := config.MustBuildFromDevelopment()
//	    // Use engine...
//	}
func MustBuildFromDevelopment() *engine.ReportEngine {
	eng, err := BuildFromDevelopment()
	if err != nil {
		panic(fmt.Sprintf("MustBuildFromDevelopment failed: %v", err))
	}
	return eng
}

// MustBuildFromTesting is like BuildFromTesting but panics on error.
// This is useful for initialization code where failure should stop the program.
//
// Returns:
//   - *engine.ReportEngine: Engine with testing configuration
//
// Panics if the engine cannot be built.
//
// Example:
//
//	func init() {
//	    engine := config.MustBuildFromTesting()
//	    // Use engine...
//	}
func MustBuildFromTesting() *engine.ReportEngine {
	eng, err := BuildFromTesting()
	if err != nil {
		panic(fmt.Sprintf("MustBuildFromTesting failed: %v", err))
	}
	return eng
}

// ValidateAndBuild validates a config and builds an engine if valid.
// This provides explicit validation before engine creation.
//
// Parameters:
//   - cfg: Configuration to validate and use for building
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//   - error: nil on success, validation or build error otherwise
//
// Example:
//
//	cfg := config.DefaultConfig()
//	cfg.Provider.Type = "postgres"
//	engine, err := config.ValidateAndBuild(cfg)
func ValidateAndBuild(cfg engine.Config) (*engine.ReportEngine, error) {
	// Explicit validation
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Build engine
	eng, err := factory.NewEngineFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build engine: %w", err)
	}

	return eng, nil
}

// LoadOrDefault attempts to load config from file, falls back to default.
// This is useful when you want config files to be optional.
//
// Parameters:
//   - path: Path to the configuration file
//
// Returns:
//   - *engine.Config: Loaded config or default config on error
//   - error: nil if loaded from file, error if using default
//
// Example:
//
//	cfg, err := config.LoadOrDefault("config.yaml")
//	if err != nil {
//	    log.Printf("Using default config: %v", err)
//	}
//	engine, _ := factory.NewEngineFromConfig(*cfg)
func LoadOrDefault(path string) (*engine.Config, error) {
	cfg, err := LoadFromFile(path)
	if err != nil {
		// Return default config
		defaultCfg := DefaultConfig()
		return &defaultCfg, fmt.Errorf("failed to load config, using default: %w", err)
	}
	return cfg, nil
}

// LoadOrDefaultWithEnv attempts to load config from file with env overrides,
// falls back to default. This combines environment variable support with
// graceful fallback to defaults.
//
// Parameters:
//   - path: Path to the configuration file
//
// Returns:
//   - *engine.Config: Loaded config or default config on error
//   - error: nil if loaded from file, error if using default
//
// Example:
//
//	cfg, err := config.LoadOrDefaultWithEnv("config.yaml")
//	if err != nil {
//	    log.Printf("Using default config: %v", err)
//	}
//	engine, _ := factory.NewEngineFromConfig(*cfg)
func LoadOrDefaultWithEnv(path string) (*engine.Config, error) {
	cfg, err := LoadFromFileWithEnv(path)
	if err != nil {
		// Return default config
		defaultCfg := DefaultConfig()
		return &defaultCfg, fmt.Errorf("failed to load config, using default: %w", err)
	}
	return cfg, nil
}

// BuildFromConfigOrFile attempts to use provided config, falls back to loading from file.
// This is useful when you want to provide config programmatically but allow
// file-based configuration as a fallback.
//
// Parameters:
//   - cfg: Optional configuration (can be nil)
//   - path: Path to configuration file (used if cfg is nil)
//
// Returns:
//   - *engine.ReportEngine: Fully configured engine ready to run
//   - error: nil on success, error otherwise
//
// Example:
//
//	var cfg *engine.Config
//	if userProvidedConfig {
//	    cfg = &userConfig
//	}
//	engine, err := config.BuildFromConfigOrFile(cfg, "config.yaml")
func BuildFromConfigOrFile(cfg *engine.Config, path string) (*engine.ReportEngine, error) {
	var finalCfg *engine.Config
	var err error

	if cfg != nil {
		// Use provided config
		finalCfg = cfg
	} else {
		// Load from file
		finalCfg, err = LoadFromFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Build engine
	eng, err := factory.NewEngineFromConfig(*finalCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build engine: %w", err)
	}

	return eng, nil
}
