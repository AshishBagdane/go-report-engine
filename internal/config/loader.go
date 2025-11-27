// Package config provides configuration loading capabilities for the report engine.
// It supports loading from YAML and JSON files with environment variable overrides,
// validation, and default values.
//
// The package is designed to handle production-grade configuration needs:
//   - Multiple file formats (YAML, JSON)
//   - Environment variable overrides
//   - Comprehensive validation
//   - Clear error messages
//   - Default values
//
// Example usage:
//
//	// Load from YAML file
//	config, err := config.LoadFromFile("config.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//
//	// Load with environment overrides
//	config, err := config.LoadFromFileWithEnv("config.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"gopkg.in/yaml.v3"
)

const (
	// EnvPrefix is the prefix for environment variables.
	// Environment variables should be in the format: ENGINE_PROVIDER_TYPE, ENGINE_FORMATTER_TYPE, etc.
	EnvPrefix = "ENGINE"
)

// Loader handles configuration loading from various sources.
type Loader struct {
	// applyEnvOverrides determines if environment variables should override file config.
	applyEnvOverrides bool
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		applyEnvOverrides: false,
	}
}

// WithEnvOverrides enables environment variable overrides.
// When enabled, environment variables prefixed with ENGINE_ will override
// values from the configuration file.
//
// Example:
//
//	ENGINE_PROVIDER_TYPE=postgres
//	ENGINE_FORMATTER_TYPE=json
//	ENGINE_OUTPUT_TYPE=file
func (l *Loader) WithEnvOverrides() *Loader {
	l.applyEnvOverrides = true
	return l
}

// LoadFromFile loads configuration from a file (YAML or JSON).
// The file format is determined by the file extension (.yaml, .yml, or .json).
//
// Parameters:
//   - path: Path to the configuration file
//
// Returns:
//   - *engine.Config: Loaded and validated configuration
//   - error: nil on success, error describing the failure otherwise
//
// Example:
//
//	config, err := loader.LoadFromFile("config.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
func (l *Loader) LoadFromFile(path string) (*engine.Config, error) {
	// Validate file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file contents
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format from extension
	ext := strings.ToLower(filepath.Ext(path))
	var config engine.Config

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	// Apply environment overrides if enabled
	if l.applyEnvOverrides {
		applyEnvironmentOverrides(&config)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadFromBytes loads configuration from raw bytes (YAML or JSON).
// The format is automatically detected based on the content.
//
// Parameters:
//   - data: Raw configuration data
//   - format: "yaml" or "json"
//
// Returns:
//   - *engine.Config: Loaded and validated configuration
//   - error: nil on success, error describing the failure otherwise
func (l *Loader) LoadFromBytes(data []byte, format string) (*engine.Config, error) {
	var config engine.Config

	switch strings.ToLower(format) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case "json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s (use 'yaml' or 'json')", format)
	}

	// Apply environment overrides if enabled
	if l.applyEnvOverrides {
		applyEnvironmentOverrides(&config)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// applyEnvironmentOverrides applies environment variable overrides to the config.
// Environment variables follow the pattern: ENGINE_<COMPONENT>_<FIELD>=value
//
// Supported overrides:
//   - ENGINE_PROVIDER_TYPE: Override provider type
//   - ENGINE_FORMATTER_TYPE: Override formatter type
//   - ENGINE_OUTPUT_TYPE: Override output type
//   - ENGINE_PROVIDER_PARAM_<KEY>: Override provider param
//   - ENGINE_FORMATTER_PARAM_<KEY>: Override formatter param
//   - ENGINE_OUTPUT_PARAM_<KEY>: Override output param
func applyEnvironmentOverrides(config *engine.Config) {
	// Provider type override
	if val := os.Getenv(EnvPrefix + "_PROVIDER_TYPE"); val != "" {
		config.Provider.Type = val
	}

	// Formatter type override
	if val := os.Getenv(EnvPrefix + "_FORMATTER_TYPE"); val != "" {
		config.Formatter.Type = val
	}

	// Output type override
	if val := os.Getenv(EnvPrefix + "_OUTPUT_TYPE"); val != "" {
		config.Output.Type = val
	}

	// Provider params overrides
	applyParamOverrides(&config.Provider.Params, EnvPrefix+"_PROVIDER_PARAM_")

	// Formatter params overrides
	applyParamOverrides(&config.Formatter.Params, EnvPrefix+"_FORMATTER_PARAM_")

	// Output params overrides
	applyParamOverrides(&config.Output.Params, EnvPrefix+"_OUTPUT_PARAM_")
}

// applyParamOverrides applies environment variable overrides to a params map.
// It looks for environment variables with the given prefix and adds/updates
// the corresponding key in the params map.
func applyParamOverrides(params *map[string]string, prefix string) {
	// Initialize map if nil
	if *params == nil {
		*params = make(map[string]string)
	}

	// Scan all environment variables
	for _, env := range os.Environ() {
		// Split key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Check if this env var matches our prefix
		if strings.HasPrefix(key, prefix) {
			// Extract the param key (lowercase)
			paramKey := strings.ToLower(strings.TrimPrefix(key, prefix))
			(*params)[paramKey] = value
		}
	}
}

// LoadFromFile is a convenience function that creates a loader and loads a config file.
func LoadFromFile(path string) (*engine.Config, error) {
	return NewLoader().LoadFromFile(path)
}

// LoadFromFileWithEnv is a convenience function that creates a loader with env overrides
// and loads a config file.
func LoadFromFileWithEnv(path string) (*engine.Config, error) {
	return NewLoader().WithEnvOverrides().LoadFromFile(path)
}

// LoadFromBytes is a convenience function that creates a loader and loads from bytes.
func LoadFromBytes(data []byte, format string) (*engine.Config, error) {
	return NewLoader().LoadFromBytes(data, format)
}
