// Package errors provides comprehensive error handling for the report engine.
// It defines custom error types for each component with context information
// and error classification for handling transient vs permanent failures.
package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorType categorizes errors for appropriate handling strategies.
type ErrorType int

const (
	// ErrorTypeUnknown represents an unclassified error.
	ErrorTypeUnknown ErrorType = iota

	// ErrorTypeValidation represents input validation failures.
	// These are permanent errors that won't succeed on retry.
	ErrorTypeValidation

	// ErrorTypeConfiguration represents configuration issues.
	// These are permanent errors requiring configuration changes.
	ErrorTypeConfiguration

	// ErrorTypeTransient represents temporary failures that may succeed on retry.
	// Examples: network timeouts, temporary resource unavailability.
	ErrorTypeTransient

	// ErrorTypePermanent represents permanent failures that won't succeed on retry.
	// Examples: invalid data format, missing required fields.
	ErrorTypePermanent

	// ErrorTypeResource represents resource-related failures.
	// Examples: disk full, memory exhausted, connection pool exhausted.
	ErrorTypeResource
)

// String returns a human-readable representation of the error type.
func (e ErrorType) String() string {
	switch e {
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeConfiguration:
		return "configuration"
	case ErrorTypeTransient:
		return "transient"
	case ErrorTypePermanent:
		return "permanent"
	case ErrorTypeResource:
		return "resource"
	default:
		return "unknown"
	}
}

// Component represents which part of the pipeline failed.
type Component string

const (
	ComponentProvider  Component = "provider"
	ComponentProcessor Component = "processor"
	ComponentFormatter Component = "formatter"
	ComponentOutput    Component = "output"
	ComponentEngine    Component = "engine"
	ComponentFactory   Component = "factory"
	ComponentRegistry  Component = "registry"
)

// EngineError is the base error type for all report engine errors.
// It provides structured context information and error classification.
type EngineError struct {
	// Component identifies which part of the pipeline failed
	Component Component

	// Operation describes what operation was being performed
	Operation string

	// ErrorType categorizes the error for handling
	Type ErrorType

	// Err is the underlying error
	Err error

	// Context provides additional information about the failure
	Context map[string]interface{}

	// Timestamp records when the error occurred
	Timestamp time.Time

	// Retryable indicates if the operation can be retried
	Retryable bool
}

// Error implements the error interface.
func (e *EngineError) Error() string {
	var sb strings.Builder

	// Component and operation
	sb.WriteString(fmt.Sprintf("[%s:%s] ", e.Component, e.Operation))

	// Error message
	if e.Err != nil {
		sb.WriteString(e.Err.Error())
	}

	// Context information
	if len(e.Context) > 0 {
		sb.WriteString(" | context: {")
		first := true
		for k, v := range e.Context {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s: %v", k, v))
			first = false
		}
		sb.WriteString("}")
	}

	// Type information
	if e.Type != ErrorTypeUnknown {
		sb.WriteString(fmt.Sprintf(" [type: %s]", e.Type))
	}

	return sb.String()
}

// Unwrap returns the underlying error for error chain traversal.
func (e *EngineError) Unwrap() error {
	return e.Err
}

// IsTransient returns true if this is a transient error that may succeed on retry.
func (e *EngineError) IsTransient() bool {
	return e.Type == ErrorTypeTransient
}

// IsPermanent returns true if this is a permanent error that won't succeed on retry.
func (e *EngineError) IsPermanent() bool {
	return e.Type == ErrorTypePermanent
}

// IsValidation returns true if this is a validation error.
func (e *EngineError) IsValidation() bool {
	return e.Type == ErrorTypeValidation
}

// IsConfiguration returns true if this is a configuration error.
func (e *EngineError) IsConfiguration() bool {
	return e.Type == ErrorTypeConfiguration
}

// IsResource returns true if this is a resource-related error.
func (e *EngineError) IsResource() bool {
	return e.Type == ErrorTypeResource
}

// NewEngineError creates a new EngineError with the given parameters.
func NewEngineError(component Component, operation string, errorType ErrorType, err error) *EngineError {
	return &EngineError{
		Component: component,
		Operation: operation,
		Type:      errorType,
		Err:       err,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Retryable: errorType == ErrorTypeTransient,
	}
}

// WithContext adds context information to the error.
func (e *EngineError) WithContext(key string, value interface{}) *EngineError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context entries to the error.
func (e *EngineError) WithContextMap(ctx map[string]interface{}) *EngineError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range ctx {
		e.Context[k] = v
	}
	return e
}

// Wrap wraps an error with component and operation context.
// If err is already an EngineError, it preserves the original context.
func Wrap(component Component, operation string, err error) error {
	if err == nil {
		return nil
	}

	// If already an EngineError, wrap it with additional context
	if engineErr, ok := err.(*EngineError); ok {
		return &EngineError{
			Component: component,
			Operation: operation,
			Type:      engineErr.Type,
			Err:       engineErr,
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
			Retryable: engineErr.Retryable,
		}
	}

	// Create new EngineError
	return NewEngineError(component, operation, ErrorTypeUnknown, err)
}

// WrapWithType wraps an error with component, operation, and type classification.
func WrapWithType(component Component, operation string, errorType ErrorType, err error) error {
	if err == nil {
		return nil
	}

	return NewEngineError(component, operation, errorType, err)
}

// IsEngineError checks if an error is an EngineError.
func IsEngineError(err error) bool {
	_, ok := err.(*EngineError)
	return ok
}

// GetErrorType returns the ErrorType of an error if it's an EngineError.
// Returns ErrorTypeUnknown for non-EngineErrors.
func GetErrorType(err error) ErrorType {
	if engineErr, ok := err.(*EngineError); ok {
		return engineErr.Type
	}
	return ErrorTypeUnknown
}

// IsRetryable returns true if the error indicates a retryable operation.
func IsRetryable(err error) bool {
	if engineErr, ok := err.(*EngineError); ok {
		return engineErr.Retryable
	}
	return false
}

// GetErrorChain returns all errors in the error chain as a slice.
func GetErrorChain(err error) []error {
	var chain []error
	for err != nil {
		chain = append(chain, err)
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrapper.Unwrap()
		} else {
			break
		}
	}
	return chain
}

// GetRootCause returns the root cause of an error by traversing the chain.
func GetRootCause(err error) error {
	for {
		if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
			underlying := unwrapper.Unwrap()
			if underlying == nil {
				return err
			}
			err = underlying
		} else {
			return err
		}
	}
}

// ErrorContext is a helper for building errors with context.
type ErrorContext struct {
	component Component
	operation string
	errorType ErrorType
	context   map[string]interface{}
}

// NewErrorContext creates a new error context builder.
func NewErrorContext(component Component, operation string) *ErrorContext {
	return &ErrorContext{
		component: component,
		operation: operation,
		errorType: ErrorTypeUnknown,
		context:   make(map[string]interface{}),
	}
}

// WithType sets the error type.
func (ec *ErrorContext) WithType(errorType ErrorType) *ErrorContext {
	ec.errorType = errorType
	return ec
}

// WithContext adds a context key-value pair.
func (ec *ErrorContext) WithContext(key string, value interface{}) *ErrorContext {
	ec.context[key] = value
	return ec
}

// Wrap wraps an error with the configured context.
func (ec *ErrorContext) Wrap(err error) error {
	if err == nil {
		return nil
	}

	engineErr := NewEngineError(ec.component, ec.operation, ec.errorType, err)
	engineErr.Context = ec.context
	return engineErr
}

// New creates a new error with the configured context.
func (ec *ErrorContext) New(message string) error {
	return ec.Wrap(fmt.Errorf("%s", message))
}
