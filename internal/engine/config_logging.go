package engine

import "github.com/AshishBagdane/go-report-engine/internal/logging"

// LoggingConfig represents logging configuration for the engine.
type LoggingConfig struct {
	// Enabled controls whether logging is enabled.
	// Default: true
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Level sets the minimum log level (debug, info, warn, error).
	// Default: info
	Level string `json:"level" yaml:"level"`

	// Format sets the output format (json or text).
	// Default: json
	Format string `json:"format" yaml:"format"`

	// AddSource includes file and line number in logs.
	// Default: false (performance overhead)
	AddSource bool `json:"add_source" yaml:"add_source"`

	// Component sets a component name for all logs.
	// Default: "engine"
	Component string `json:"component" yaml:"component"`
}

// DefaultLoggingConfig returns a production-ready logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Enabled:   true,
		Level:     "info",
		Format:    "json",
		AddSource: false,
		Component: "engine",
	}
}

// ToLoggerConfig converts engine LoggingConfig to logging.Config.
// It validates the configuration and returns appropriate logging.Config.
func (c LoggingConfig) ToLoggerConfig() logging.Config {
	config := logging.DefaultConfig()

	// Set component
	config.Component = c.Component

	// Parse level
	switch c.Level {
	case "debug":
		config.Level = logging.LevelDebug
	case "info":
		config.Level = logging.LevelInfo
	case "warn", "warning":
		config.Level = logging.LevelWarn
	case "error":
		config.Level = logging.LevelError
	default:
		config.Level = logging.LevelInfo // Default to info on invalid
	}

	// Parse format
	switch c.Format {
	case "json":
		config.Format = logging.FormatJSON
	case "text":
		config.Format = logging.FormatText
	default:
		config.Format = logging.FormatJSON // Default to JSON on invalid
	}

	// Set source
	config.AddSource = c.AddSource

	return config
}

// Validate validates the logging configuration.
// Returns nil if valid, error describing the issue if invalid.
func (c LoggingConfig) Validate() error {
	// Level validation
	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warn":    true,
		"warning": true,
		"error":   true,
	}

	if c.Level != "" && !validLevels[c.Level] {
		return ErrInvalidLogLevel
	}

	// Format validation
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if c.Format != "" && !validFormats[c.Format] {
		return ErrInvalidLogFormat
	}

	return nil
}
