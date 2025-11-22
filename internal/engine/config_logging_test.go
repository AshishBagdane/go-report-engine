package engine

import (
	"testing"

	"github.com/AshishBagdane/report-engine/internal/logging"
)

// TestDefaultLoggingConfig tests default logging configuration
func TestDefaultLoggingConfig(t *testing.T) {
	config := DefaultLoggingConfig()

	if !config.Enabled {
		t.Error("Default logging should be enabled")
	}

	if config.Level != "info" {
		t.Errorf("Default level = %q, expected 'info'", config.Level)
	}

	if config.Format != "json" {
		t.Errorf("Default format = %q, expected 'json'", config.Format)
	}

	if config.AddSource {
		t.Error("Default AddSource should be false")
	}

	if config.Component != "engine" {
		t.Errorf("Default component = %q, expected 'engine'", config.Component)
	}
}

// TestLoggingConfigToLoggerConfig tests conversion to logger config
func TestLoggingConfigToLoggerConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         LoggingConfig
		expectedLevel  logging.Level
		expectedFormat logging.Format
	}{
		{
			name: "debug level",
			config: LoggingConfig{
				Level:  "debug",
				Format: "json",
			},
			expectedLevel:  logging.LevelDebug,
			expectedFormat: logging.FormatJSON,
		},
		{
			name: "info level",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
			expectedLevel:  logging.LevelInfo,
			expectedFormat: logging.FormatJSON,
		},
		{
			name: "warn level",
			config: LoggingConfig{
				Level:  "warn",
				Format: "text",
			},
			expectedLevel:  logging.LevelWarn,
			expectedFormat: logging.FormatText,
		},
		{
			name: "warning level alias",
			config: LoggingConfig{
				Level:  "warning",
				Format: "json",
			},
			expectedLevel:  logging.LevelWarn,
			expectedFormat: logging.FormatJSON,
		},
		{
			name: "error level",
			config: LoggingConfig{
				Level:  "error",
				Format: "text",
			},
			expectedLevel:  logging.LevelError,
			expectedFormat: logging.FormatText,
		},
		{
			name: "invalid level defaults to info",
			config: LoggingConfig{
				Level:  "invalid",
				Format: "json",
			},
			expectedLevel:  logging.LevelInfo,
			expectedFormat: logging.FormatJSON,
		},
		{
			name: "invalid format defaults to json",
			config: LoggingConfig{
				Level:  "info",
				Format: "invalid",
			},
			expectedLevel:  logging.LevelInfo,
			expectedFormat: logging.FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerConfig := tt.config.ToLoggerConfig()

			if loggerConfig.Level != tt.expectedLevel {
				t.Errorf("Level = %v, expected %v", loggerConfig.Level, tt.expectedLevel)
			}

			if loggerConfig.Format != tt.expectedFormat {
				t.Errorf("Format = %v, expected %v", loggerConfig.Format, tt.expectedFormat)
			}
		})
	}
}

// TestLoggingConfigComponent tests component name conversion
func TestLoggingConfigComponent(t *testing.T) {
	config := LoggingConfig{
		Level:     "info",
		Format:    "json",
		Component: "test-component",
	}

	loggerConfig := config.ToLoggerConfig()

	if loggerConfig.Component != "test-component" {
		t.Errorf("Component = %q, expected 'test-component'", loggerConfig.Component)
	}
}

// TestLoggingConfigAddSource tests AddSource conversion
func TestLoggingConfigAddSource(t *testing.T) {
	tests := []struct {
		name      string
		addSource bool
	}{
		{"add source true", true},
		{"add source false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggingConfig{
				Level:     "info",
				Format:    "json",
				AddSource: tt.addSource,
			}

			loggerConfig := config.ToLoggerConfig()

			if loggerConfig.AddSource != tt.addSource {
				t.Errorf("AddSource = %v, expected %v", loggerConfig.AddSource, tt.addSource)
			}
		})
	}
}

// TestLoggingConfigValidate tests validation
func TestLoggingConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      LoggingConfig
		shouldError bool
		expectedErr error
	}{
		{
			name: "valid debug level",
			config: LoggingConfig{
				Level:  "debug",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "valid info level",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "valid warn level",
			config: LoggingConfig{
				Level:  "warn",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "valid warning level",
			config: LoggingConfig{
				Level:  "warning",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "valid error level",
			config: LoggingConfig{
				Level:  "error",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "valid json format",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "valid text format",
			config: LoggingConfig{
				Level:  "info",
				Format: "text",
			},
			shouldError: false,
		},
		{
			name: "invalid level",
			config: LoggingConfig{
				Level:  "invalid",
				Format: "json",
			},
			shouldError: true,
			expectedErr: ErrInvalidLogLevel,
		},
		{
			name: "invalid format",
			config: LoggingConfig{
				Level:  "info",
				Format: "invalid",
			},
			shouldError: true,
			expectedErr: ErrInvalidLogFormat,
		},
		{
			name: "empty level is valid",
			config: LoggingConfig{
				Level:  "",
				Format: "json",
			},
			shouldError: false,
		},
		{
			name: "empty format is valid",
			config: LoggingConfig{
				Level:  "info",
				Format: "",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.shouldError {
				if err == nil {
					t.Error("Validate() should return error")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Error = %v, expected %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() should not return error, got: %v", err)
				}
			}
		})
	}
}

// TestLoggingConfigCaseSensitivity tests case sensitivity
func TestLoggingConfigCaseSensitivity(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		shouldError bool
	}{
		{"lowercase debug", "debug", false},
		{"uppercase DEBUG", "DEBUG", true}, // Should be invalid
		{"mixed case Debug", "Debug", true},
		{"lowercase info", "info", false},
		{"uppercase INFO", "INFO", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggingConfig{
				Level:  tt.level,
				Format: "json",
			}

			err := config.Validate()

			if tt.shouldError && err == nil {
				t.Error("Validate() should return error for invalid case")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Validate() should not return error, got: %v", err)
			}
		})
	}
}

// TestLoggingConfigEnabled tests enabled flag
func TestLoggingConfigEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled true", true},
		{"enabled false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggingConfig{
				Enabled: tt.enabled,
				Level:   "info",
				Format:  "json",
			}

			if config.Enabled != tt.enabled {
				t.Errorf("Enabled = %v, expected %v", config.Enabled, tt.enabled)
			}
		})
	}
}

// TestLoggingConfigCompleteConversion tests full configuration conversion
func TestLoggingConfigCompleteConversion(t *testing.T) {
	engineConfig := LoggingConfig{
		Enabled:   true,
		Level:     "debug",
		Format:    "text",
		AddSource: true,
		Component: "my-component",
	}

	loggerConfig := engineConfig.ToLoggerConfig()

	// Verify all fields converted correctly
	if loggerConfig.Level != logging.LevelDebug {
		t.Error("Level not converted correctly")
	}
	if loggerConfig.Format != logging.FormatText {
		t.Error("Format not converted correctly")
	}
	if !loggerConfig.AddSource {
		t.Error("AddSource not converted correctly")
	}
	if loggerConfig.Component != "my-component" {
		t.Error("Component not converted correctly")
	}
}

// TestLoggingConfigZeroValue tests zero value behavior
func TestLoggingConfigZeroValue(t *testing.T) {
	var config LoggingConfig

	// Zero value should be valid
	err := config.Validate()
	if err != nil {
		t.Errorf("Zero value should be valid, got: %v", err)
	}

	// Should convert without panic
	loggerConfig := config.ToLoggerConfig()
	if loggerConfig.Level != logging.LevelInfo {
		t.Error("Zero value should default to Info level")
	}
	if loggerConfig.Format != logging.FormatJSON {
		t.Error("Zero value should default to JSON format")
	}
}

// TestErrInvalidLogLevel tests error constant
func TestErrInvalidLogLevel(t *testing.T) {
	if ErrInvalidLogLevel == nil {
		t.Error("ErrInvalidLogLevel should not be nil")
	}
}

// TestErrInvalidLogFormat tests error constant
func TestErrInvalidLogFormat(t *testing.T) {
	if ErrInvalidLogFormat == nil {
		t.Error("ErrInvalidLogFormat should not be nil")
	}
}

// BenchmarkToLoggerConfig benchmarks config conversion
func BenchmarkToLoggerConfig(b *testing.B) {
	config := LoggingConfig{
		Enabled:   true,
		Level:     "info",
		Format:    "json",
		AddSource: false,
		Component: "engine",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ToLoggerConfig()
	}
}

// BenchmarkValidate benchmarks validation
func BenchmarkValidate(b *testing.B) {
	config := LoggingConfig{
		Enabled:   true,
		Level:     "info",
		Format:    "json",
		AddSource: false,
		Component: "engine",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}
