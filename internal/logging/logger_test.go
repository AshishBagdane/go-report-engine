package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// TestNewLogger tests logger creation with various configurations
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "default config",
			config: Config{
				Level:     LevelInfo,
				Format:    FormatJSON,
				Component: "test",
			},
		},
		{
			name: "debug level",
			config: Config{
				Level:     LevelDebug,
				Format:    FormatText,
				Component: "debug",
			},
		},
		{
			name: "no component",
			config: Config{
				Level:  LevelWarn,
				Format: FormatJSON,
			},
		},
		{
			name: "with source",
			config: Config{
				Level:     LevelError,
				Format:    FormatJSON,
				Component: "source",
				AddSource: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.config.Output = &buf

			logger := NewLogger(tt.config)

			if logger == nil {
				t.Fatal("NewLogger() returned nil")
			}

			if logger.component != tt.config.Component {
				t.Errorf("component = %q, expected %q", logger.component, tt.config.Component)
			}
		})
	}
}

// TestNew tests default logger creation
func TestNew(t *testing.T) {
	logger := New()

	if logger == nil {
		t.Fatal("New() returned nil")
	}
}

// TestDefaultConfig tests default configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != LevelInfo {
		t.Errorf("Default level = %v, expected %v", config.Level, LevelInfo)
	}

	if config.Format != FormatJSON {
		t.Errorf("Default format = %v, expected %v", config.Format, FormatJSON)
	}

	if config.Output != nil {
		t.Error("Default output should be nil (will use os.Stderr)")
	}

	if len(config.SensitiveKeys) == 0 {
		t.Error("Default config should have sensitive keys defined")
	}
}

// TestLevelString tests Level.String() method
func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Level.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestFormatString tests Format.String() method
func TestFormatString(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{FormatJSON, "json"},
		{FormatText, "text"},
		{Format(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.format.String()
			if result != tt.expected {
				t.Errorf("Format.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestLoggerDebug tests Debug logging
func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     LevelDebug,
		Format:    FormatJSON,
		Output:    &buf,
		Component: "test",
	})

	logger.Debug("debug message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Output should contain debug message")
	}
	if !strings.Contains(output, "key") {
		t.Error("Output should contain key field")
	}
}

// TestLoggerInfo tests Info logging
func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     LevelInfo,
		Format:    FormatJSON,
		Output:    &buf,
		Component: "test",
	})

	logger.Info("info message", "count", 42)

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Error("Output should contain info message")
	}
	if !strings.Contains(output, "42") {
		t.Error("Output should contain count value")
	}
}

// TestLoggerWarn tests Warn logging
func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     LevelWarn,
		Format:    FormatJSON,
		Output:    &buf,
		Component: "test",
	})

	logger.Warn("warning message", "threshold", 100)

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Error("Output should contain warning message")
	}
}

// TestLoggerError tests Error logging
func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     LevelError,
		Format:    FormatJSON,
		Output:    &buf,
		Component: "test",
	})

	logger.Error("error message", "error", "failed")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Error("Output should contain error message")
	}
}

// TestLoggerLevelFiltering tests that log levels are properly filtered
func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelWarn, // Only Warn and Error should be logged
		Format: FormatJSON,
		Output: &buf,
	})

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged at Warn level")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not be logged at Warn level")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be logged at Warn level")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be logged at Warn level")
	}
}

// TestLoggerWithComponent tests WithComponent method
func TestLoggerWithComponent(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	componentLogger := logger.WithComponent("engine")
	componentLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "engine") {
		t.Error("Output should contain component name")
	}
}

// TestLoggerWith tests With method
func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	contextLogger := logger.With("request_id", "req-123", "user_id", "user-456")
	contextLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "req-123") {
		t.Error("Output should contain request_id")
	}
	if !strings.Contains(output, "user-456") {
		t.Error("Output should contain user_id")
	}
}

// TestLoggerContextMethods tests context-aware logging methods
func TestLoggerContextMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelDebug,
		Format: FormatJSON,
		Output: &buf,
	})

	ctx := WithRequestID(context.Background(), "req-abc")
	ctx = WithCorrelationID(ctx, "corr-xyz")

	logger.DebugContext(ctx, "debug with context")
	logger.InfoContext(ctx, "info with context")
	logger.WarnContext(ctx, "warn with context")
	logger.ErrorContext(ctx, "error with context")

	output := buf.String()

	// Should contain request_id and correlation_id in all messages
	if strings.Count(output, "req-abc") != 4 {
		t.Errorf("Output should contain request_id 4 times")
	}
	if strings.Count(output, "corr-xyz") != 4 {
		t.Errorf("Output should contain correlation_id 4 times")
	}
}

// TestSensitiveDataFiltering tests that sensitive fields are redacted
func TestSensitiveDataFiltering(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		shouldRedact bool
	}{
		{"password", "password", "secret123", true},
		{"api_key", "api_key", "key-abc-123", true},
		{"token", "token", "bearer-xyz", true},
		{"secret", "secret", "my-secret", true},
		{"authorization", "authorization", "Bearer token", true},
		{"normal field", "username", "alice", false},
		{"case insensitive", "PASSWORD", "secret", true},
		{"case insensitive 2", "API_KEY", "key", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(Config{
				Level:  LevelInfo,
				Format: FormatJSON,
				Output: &buf,
			})

			logger.Info("test", tt.key, tt.value)

			output := buf.String()
			if tt.shouldRedact {
				if strings.Contains(output, tt.value) {
					t.Errorf("Sensitive value %q should be redacted", tt.value)
				}
				if !strings.Contains(output, "[REDACTED]") {
					t.Error("Output should contain [REDACTED]")
				}
			} else {
				if !strings.Contains(output, tt.value) {
					t.Errorf("Non-sensitive value %q should not be redacted", tt.value)
				}
			}
		})
	}
}

// TestCustomSensitiveKeys tests custom sensitive keys configuration
func TestCustomSensitiveKeys(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:         LevelInfo,
		Format:        FormatJSON,
		Output:        &buf,
		SensitiveKeys: []string{"custom_secret", "internal_token"},
	})

	logger.Info("test", "custom_secret", "should-be-redacted")

	output := buf.String()
	if strings.Contains(output, "should-be-redacted") {
		t.Error("Custom sensitive key should be redacted")
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("Output should contain [REDACTED]")
	}
}

// TestLoggerEnabled tests the Enabled method
func TestLoggerEnabled(t *testing.T) {
	tests := []struct {
		name            string
		loggerLevel     Level
		checkLevel      Level
		shouldBeEnabled bool
	}{
		{"debug logger, check debug", LevelDebug, LevelDebug, true},
		{"debug logger, check info", LevelDebug, LevelInfo, true},
		{"info logger, check debug", LevelInfo, LevelDebug, false},
		{"info logger, check info", LevelInfo, LevelInfo, true},
		{"warn logger, check info", LevelWarn, LevelInfo, false},
		{"warn logger, check error", LevelWarn, LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(Config{
				Level:  tt.loggerLevel,
				Format: FormatJSON,
				Output: &buf,
			})

			enabled := logger.Enabled(tt.checkLevel)
			if enabled != tt.shouldBeEnabled {
				t.Errorf("Enabled() = %v, expected %v", enabled, tt.shouldBeEnabled)
			}
		})
	}
}

// TestJSONFormat tests JSON output format
func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:     LevelInfo,
		Format:    FormatJSON,
		Output:    &buf,
		Component: "test",
	})

	logger.Info("json test", "key", "value", "count", 42)

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Check expected fields
	if result["msg"] != "json test" {
		t.Error("JSON should contain msg field")
	}
	if result["component"] != "test" {
		t.Error("JSON should contain component field")
	}
	if result["key"] != "value" {
		t.Error("JSON should contain key field")
	}
}

// TestTextFormat tests text output format
func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatText,
		Output: &buf,
	})

	logger.Info("text test", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "text test") {
		t.Error("Text output should contain message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Text output should contain key=value pair")
	}
}

// TestGlobalLogger tests global logger functions
func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	SetGlobalLogger(logger)

	if GetGlobalLogger() != logger {
		t.Error("GetGlobalLogger should return the logger we set")
	}

	// Test package-level functions
	Info("global info")
	Warn("global warn")
	Error("global error")

	output := buf.String()
	if !strings.Contains(output, "global info") {
		t.Error("Global Info should work")
	}
	if !strings.Contains(output, "global warn") {
		t.Error("Global Warn should work")
	}
	if !strings.Contains(output, "global error") {
		t.Error("Global Error should work")
	}
}

// TestGlobalLoggerContext tests global logger context functions
func TestGlobalLoggerContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	SetGlobalLogger(logger)

	ctx := WithRequestID(context.Background(), "global-req")

	InfoContext(ctx, "global info with context")
	WarnContext(ctx, "global warn with context")
	ErrorContext(ctx, "global error with context")

	output := buf.String()
	if strings.Count(output, "global-req") != 3 {
		t.Error("Global context functions should include request_id")
	}
}

// TestFilterSensitiveOddArguments tests filtering with odd number of arguments
func TestFilterSensitiveOddArguments(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	// Odd number of arguments (should not panic)
	logger.Info("test", "key", "value", "orphan")

	output := buf.String()
	if !strings.Contains(output, "test") {
		t.Error("Should handle odd number of arguments gracefully")
	}
}

// TestNilOutput tests that nil output defaults to stderr
func TestNilOutput(t *testing.T) {
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: nil, // Should default to os.Stderr
	})

	// Should not panic
	logger.Info("test message")
}

// TestLoggerConcurrentAccess tests concurrent logging
func TestLoggerConcurrentAccess(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			logger.Info("concurrent message", "goroutine", id)
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Should not panic and should contain all messages
	output := buf.String()
	messageCount := strings.Count(output, "concurrent message")
	if messageCount != goroutines {
		t.Errorf("Expected %d messages, got %d", goroutines, messageCount)
	}
}

// BenchmarkLoggerInfo benchmarks Info logging
func BenchmarkLoggerInfo(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "key", "value")
	}
}

// BenchmarkLoggerInfoWithSensitiveFilter benchmarks with sensitive filtering
func BenchmarkLoggerInfoWithSensitiveFilter(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark", "password", "secret", "normal", "value")
	}
}

// BenchmarkLoggerDisabled benchmarks when logging is disabled
func BenchmarkLoggerDisabled(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelError, // Debug won't be logged
		Format: FormatJSON,
		Output: &buf,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("benchmark message", "iteration", i)
	}
}

// BenchmarkLoggerWithContext benchmarks context logging
func BenchmarkLoggerWithContext(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	ctx := WithRequestID(context.Background(), "bench-req")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "benchmark message", "iteration", i)
	}
}

// BenchmarkJSONFormat benchmarks JSON formatting
func BenchmarkJSONFormat(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("message", "key", "value")
	}
}

// BenchmarkTextFormat benchmarks text formatting
func BenchmarkTextFormat(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger(Config{
		Level:  LevelInfo,
		Format: FormatText,
		Output: &buf,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("message", "key", "value")
	}
}
