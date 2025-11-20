// Package errors provides output-specific error types and helpers.
package errors

import (
	"fmt"
)

// OutputError represents errors that occur during output delivery.
type OutputError struct {
	*EngineError
	OutputType  string
	Destination string
	DataSize    int64
}

// NewOutputError creates a new output error.
func NewOutputError(operation string, errorType ErrorType, err error) *OutputError {
	return &OutputError{
		EngineError: NewEngineError(ComponentOutput, operation, errorType, err),
		DataSize:    -1, // -1 indicates not set
	}
}

// WithOutputType sets the output type (e.g., "console", "file", "s3", "slack").
func (e *OutputError) WithOutputType(outputType string) *OutputError {
	e.OutputType = outputType
	e.EngineError.WithContext("output_type", outputType)
	return e
}

// WithDestination sets the output destination.
func (e *OutputError) WithDestination(destination string) *OutputError {
	e.Destination = destination
	// Sanitize destination if it's a URL or path (might contain sensitive info)
	if len(destination) > 100 {
		e.EngineError.WithContext("destination_preview", destination[:100]+"...")
	} else {
		e.EngineError.WithContext("destination", destination)
	}
	return e
}

// WithDataSize sets the size of data being sent.
func (e *OutputError) WithDataSize(size int64) *OutputError {
	e.DataSize = size
	e.EngineError.WithContext("data_size", size)
	return e
}

// WithContext adds context information to the error and returns the OutputError.
func (e *OutputError) WithContext(key string, value interface{}) *OutputError {
	e.EngineError.WithContext(key, value)
	return e
}

// WithContextMap adds multiple context entries to the error and returns the OutputError.
func (e *OutputError) WithContextMap(ctx map[string]interface{}) *OutputError {
	e.EngineError.WithContextMap(ctx)
	return e
}

// Common output error constructors

// ErrOutputConnection creates an error for connection failures.
func ErrOutputConnection(outputType string, destination string, err error) *OutputError {
	return NewOutputError("connect", ErrorTypeTransient, err).
		WithOutputType(outputType).
		WithDestination(destination)
}

// ErrOutputAuthentication creates an error for authentication failures.
func ErrOutputAuthentication(outputType string, destination string, err error) *OutputError {
	return NewOutputError("authenticate", ErrorTypeConfiguration, err).
		WithOutputType(outputType).
		WithDestination(destination)
}

// ErrOutputWrite creates an error for write failures.
func ErrOutputWrite(outputType string, destination string, err error) *OutputError {
	return NewOutputError("write", ErrorTypeTransient, err).
		WithOutputType(outputType).
		WithDestination(destination)
}

// ErrOutputPermission creates an error for permission failures.
func ErrOutputPermission(outputType string, destination string, err error) *OutputError {
	return NewOutputError("write", ErrorTypeConfiguration, err).
		WithOutputType(outputType).
		WithDestination(destination)
}

// ErrOutputDiskFull creates an error when disk is full.
func ErrOutputDiskFull(outputType string, destination string, dataSize int64) *OutputError {
	return NewOutputError("write", ErrorTypeResource,
		fmt.Errorf("insufficient disk space for %d bytes", dataSize)).
		WithOutputType(outputType).
		WithDestination(destination).
		WithDataSize(dataSize)
}

// ErrOutputTimeout creates an error for timeout failures.
func ErrOutputTimeout(outputType string, destination string, err error) *OutputError {
	return NewOutputError("send", ErrorTypeTransient, err).
		WithOutputType(outputType).
		WithDestination(destination)
}

// ErrOutputSizeLimitExceeded creates an error when data size exceeds limits.
func ErrOutputSizeLimitExceeded(outputType string, size int64, limit int64) *OutputError {
	return NewOutputError("send", ErrorTypeResource,
		fmt.Errorf("data size %d bytes exceeds limit of %d bytes", size, limit)).
		WithOutputType(outputType).
		WithDataSize(size).
		WithContext("size_limit", limit)
}

// ErrOutputNotFound creates an error when output destination doesn't exist.
func ErrOutputNotFound(outputType string, destination string) *OutputError {
	return NewOutputError("send", ErrorTypePermanent,
		fmt.Errorf("destination not found: %s", destination)).
		WithOutputType(outputType).
		WithDestination(destination)
}

// ErrOutputConfiguration creates an error for configuration issues.
func ErrOutputConfiguration(outputType string, paramName string, err error) *OutputError {
	return NewOutputError("configure", ErrorTypeConfiguration, err).
		WithOutputType(outputType).
		WithContext("parameter", paramName)
}

// ErrOutputRateLimitExceeded creates an error when rate limit is exceeded.
func ErrOutputRateLimitExceeded(outputType string, destination string, limit string) *OutputError {
	return NewOutputError("send", ErrorTypeTransient,
		fmt.Errorf("rate limit exceeded: %s", limit)).
		WithOutputType(outputType).
		WithDestination(destination).
		WithContext("rate_limit", limit)
}
