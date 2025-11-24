// Package logging provides structured logging capabilities for the report engine
// using Go's standard log/slog package. It supports multiple log levels, context
// propagation, sensitive data filtering, and component-specific loggers.
//
// The package is designed for production use with:
//   - Configurable log levels (Debug, Info, Warn, Error)
//   - Multiple output formats (JSON, Text)
//   - Automatic sensitive data redaction
//   - Context-aware logging with correlation IDs
//   - Component-specific logger instances
//
// Example usage:
//
//	// Create a logger for a component
//	logger := logging.NewLogger(logging.Config{
//	    Level:     logging.LevelInfo,
//	    Format:    logging.FormatJSON,
//	    Component: "engine",
//	})
//
//	// Log with context
//	ctx := logging.WithRequestID(context.Background(), "req-123")
//	logger.InfoContext(ctx, "processing started",
//	    "record_count", 100,
//	    "provider", "postgres",
//	)
//
//	// Log errors with structured data
//	logger.Error("processing failed",
//	    "error", err,
//	    "stage", "validation",
//	    "record_index", 42,
//	)
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level represents the severity of a log message.
type Level int

const (
	// LevelDebug logs detailed information for debugging.
	// Use for development and troubleshooting.
	LevelDebug Level = iota

	// LevelInfo logs informational messages about normal operation.
	// Use for general application flow and milestones.
	LevelInfo

	// LevelWarn logs warning messages about potential issues.
	// Use for recoverable errors and unexpected conditions.
	LevelWarn

	// LevelError logs error messages about failures.
	// Use for errors that prevent normal operation.
	LevelError
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// toSlogLevel converts our Level to slog.Level.
func (l Level) toSlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Format represents the output format for log messages.
type Format int

const (
	// FormatJSON outputs logs as JSON objects.
	// Recommended for production environments.
	FormatJSON Format = iota

	// FormatText outputs logs as human-readable text.
	// Recommended for development and debugging.
	FormatText
)

// String returns the string representation of the format.
func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatText:
		return "text"
	default:
		return "unknown"
	}
}

// Config holds configuration for a Logger instance.
type Config struct {
	// Level sets the minimum log level. Messages below this level are discarded.
	Level Level

	// Format determines the output format (JSON or Text).
	Format Format

	// Output is where log messages are written. Defaults to os.Stderr.
	Output io.Writer

	// Component is the name of the component using this logger.
	// It's automatically added to all log messages as "component" field.
	Component string

	// AddSource includes file and line number in log messages when true.
	// Useful for debugging but adds overhead.
	AddSource bool

	// SensitiveKeys is a list of field names that should be redacted.
	// Values for these keys will be replaced with "[REDACTED]".
	SensitiveKeys []string
}

// DefaultConfig returns a production-ready configuration.
func DefaultConfig() Config {
	return Config{
		Level:         LevelInfo,
		Format:        FormatJSON,
		Output:        nil, // Will default to os.Stderr in NewLogger
		Component:     "",
		AddSource:     false,
		SensitiveKeys: defaultSensitiveKeys(),
	}
}

// defaultSensitiveKeys returns commonly sensitive field names.
func defaultSensitiveKeys() []string {
	return []string{
		"password",
		"secret",
		"token",
		"api_key",
		"apikey",
		"access_token",
		"refresh_token",
		"private_key",
		"credential",
		"auth",
		"authorization",
		"ssn",
		"social_security",
		"credit_card",
		"card_number",
	}
}

// Logger wraps slog.Logger with additional functionality for the report engine.
type Logger struct {
	slog          *slog.Logger
	component     string
	sensitiveKeys map[string]struct{} // For O(1) lookup
}

// NewLogger creates a new Logger with the given configuration.
// If config.Output is nil, it defaults to os.Stderr.
// If config.Component is empty, no component field is added to logs.
//
// Example:
//
//	logger := logging.NewLogger(logging.Config{
//	    Level:     logging.LevelInfo,
//	    Format:    logging.FormatJSON,
//	    Component: "provider",
//	})
func NewLogger(config Config) *Logger {
	// Set defaults
	if config.Output == nil {
		config.Output = os.Stderr
	}

	// Use default sensitive keys if none provided
	if config.SensitiveKeys == nil {
		config.SensitiveKeys = defaultSensitiveKeys()
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     config.Level.toSlogLevel(),
		AddSource: config.AddSource,
	}

	// Create appropriate handler based on format
	var handler slog.Handler
	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(config.Output, opts)
	case FormatText:
		handler = slog.NewTextHandler(config.Output, opts)
	default:
		handler = slog.NewJSONHandler(config.Output, opts)
	}

	// Build sensitive keys map for efficient lookup
	sensitiveMap := make(map[string]struct{}, len(config.SensitiveKeys))
	for _, key := range config.SensitiveKeys {
		sensitiveMap[strings.ToLower(key)] = struct{}{}
	}

	// Create base logger
	baseLogger := slog.New(handler)

	// Add component field if specified
	if config.Component != "" {
		baseLogger = baseLogger.With("component", config.Component)
	}

	return &Logger{
		slog:          baseLogger,
		component:     config.Component,
		sensitiveKeys: sensitiveMap,
	}
}

// New creates a Logger with default configuration.
// Equivalent to NewLogger(DefaultConfig()).
func New() *Logger {
	return NewLogger(DefaultConfig())
}

// WithComponent creates a new Logger with the specified component name.
// The component name is added as a field to all log messages.
//
// Example:
//
//	engineLogger := logger.WithComponent("engine")
//	engineLogger.Info("engine started")  // Will include "component": "engine"
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		slog:          l.slog.With("component", component),
		component:     component,
		sensitiveKeys: l.sensitiveKeys,
	}
}

// With creates a new Logger with the given attributes added to all log messages.
// This is useful for adding request-scoped or operation-scoped context.
//
// Example:
//
//	requestLogger := logger.With("request_id", "req-123", "user_id", "user-456")
//	requestLogger.Info("processing request")  // Will include both IDs
func (l *Logger) With(args ...interface{}) *Logger {
	// Filter sensitive data before adding to logger
	filtered := l.filterSensitive(args...)

	return &Logger{
		slog:          l.slog.With(filtered...),
		component:     l.component,
		sensitiveKeys: l.sensitiveKeys,
	}
}

// Debug logs a debug-level message.
// Debug logs are typically used during development and troubleshooting.
// They are not logged in production unless explicitly configured.
//
// Example:
//
//	logger.Debug("processing record",
//	    "record_id", 123,
//	    "fields", map[string]interface{}{"name": "Alice"},
//	)
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.slog.Debug(msg, l.filterSensitive(args...)...)
}

// DebugContext logs a debug-level message with context.
// Context is used to extract request IDs and other contextual information.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	args = l.addContextFields(ctx, args...)
	l.slog.DebugContext(ctx, msg, l.filterSensitive(args...)...)
}

// Info logs an info-level message.
// Info logs represent normal operation milestones and important events.
//
// Example:
//
//	logger.Info("report generated successfully",
//	    "record_count", 1000,
//	    "duration_ms", 250,
//	)
func (l *Logger) Info(msg string, args ...interface{}) {
	l.slog.Info(msg, l.filterSensitive(args...)...)
}

// InfoContext logs an info-level message with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	args = l.addContextFields(ctx, args...)
	l.slog.InfoContext(ctx, msg, l.filterSensitive(args...)...)
}

// Warn logs a warning-level message.
// Warnings indicate potential issues or unexpected conditions that don't
// prevent the system from functioning.
//
// Example:
//
//	logger.Warn("provider connection slow",
//	    "latency_ms", 5000,
//	    "threshold_ms", 1000,
//	)
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.slog.Warn(msg, l.filterSensitive(args...)...)
}

// WarnContext logs a warning-level message with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	args = l.addContextFields(ctx, args...)
	l.slog.WarnContext(ctx, msg, l.filterSensitive(args...)...)
}

// Error logs an error-level message.
// Error logs indicate failures that prevent normal operation.
//
// Example:
//
//	logger.Error("validation failed",
//	    "error", err,
//	    "record_index", 42,
//	    "field", "email",
//	)
func (l *Logger) Error(msg string, args ...interface{}) {
	l.slog.Error(msg, l.filterSensitive(args...)...)
}

// ErrorContext logs an error-level message with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	args = l.addContextFields(ctx, args...)
	l.slog.ErrorContext(ctx, msg, l.filterSensitive(args...)...)
}

// filterSensitive redacts sensitive field values from log arguments.
// It scans through key-value pairs and replaces values for sensitive keys.
func (l *Logger) filterSensitive(args ...interface{}) []interface{} {
	if len(l.sensitiveKeys) == 0 {
		return args
	}

	// Process pairs of key-value arguments
	filtered := make([]interface{}, len(args))
	copy(filtered, args)

	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			// Check if key is sensitive (case-insensitive)
			if _, isSensitive := l.sensitiveKeys[strings.ToLower(key)]; isSensitive {
				// Redact the value at position i+1
				filtered[i+1] = "[REDACTED]"
			}
		}
	}

	return filtered
}

// addContextFields extracts fields from context and adds them to args.
// Currently supports request_id and correlation_id from context.
func (l *Logger) addContextFields(ctx context.Context, args ...interface{}) []interface{} {
	// Extract request ID if present
	if requestID := GetRequestID(ctx); requestID != "" {
		args = append(args, "request_id", requestID)
	}

	// Extract correlation ID if present
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		args = append(args, "correlation_id", correlationID)
	}

	return args
}

// Enabled reports whether the logger is enabled at the given level.
// This can be used to avoid expensive operations when logging is disabled.
//
// Example:
//
//	if logger.Enabled(logging.LevelDebug) {
//	    expensiveData := computeExpensiveDebugData()
//	    logger.Debug("debug info", "data", expensiveData)
//	}
func (l *Logger) Enabled(level Level) bool {
	return l.slog.Enabled(context.Background(), level.toSlogLevel())
}

// --- Global Logger ---

var globalLogger = New()

// SetGlobalLogger sets the global logger instance.
// This is used by package-level logging functions.
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the current global logger.
func GetGlobalLogger() *Logger {
	return globalLogger
}

// --- Package-level logging functions ---

// Debug logs a debug-level message using the global logger.
func Debug(msg string, args ...interface{}) {
	globalLogger.Debug(msg, args...)
}

// DebugContext logs a debug-level message with context using the global logger.
func DebugContext(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.DebugContext(ctx, msg, args...)
}

// Info logs an info-level message using the global logger.
func Info(msg string, args ...interface{}) {
	globalLogger.Info(msg, args...)
}

// InfoContext logs an info-level message with context using the global logger.
func InfoContext(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.InfoContext(ctx, msg, args...)
}

// Warn logs a warning-level message using the global logger.
func Warn(msg string, args ...interface{}) {
	globalLogger.Warn(msg, args...)
}

// WarnContext logs a warning-level message with context using the global logger.
func WarnContext(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.WarnContext(ctx, msg, args...)
}

// Error logs an error-level message using the global logger.
func Error(msg string, args ...interface{}) {
	globalLogger.Error(msg, args...)
}

// ErrorContext logs an error-level message with context using the global logger.
func ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.ErrorContext(ctx, msg, args...)
}
