// Package errors provides processor-specific error types and helpers.
package errors

import (
	"fmt"
)

// ProcessorError represents errors that occur during data processing.
type ProcessorError struct {
	*EngineError
	ProcessorType string
	ProcessorName string
	RecordIndex   int
	FieldName     string
}

// NewProcessorError creates a new processor error.
func NewProcessorError(operation string, errorType ErrorType, err error) *ProcessorError {
	return &ProcessorError{
		EngineError: NewEngineError(ComponentProcessor, operation, errorType, err),
		RecordIndex: -1, // -1 indicates not set
	}
}

// WithProcessorType sets the processor type (e.g., "filter", "validator", "transformer").
func (e *ProcessorError) WithProcessorType(processorType string) *ProcessorError {
	e.ProcessorType = processorType
	_ = e.EngineError.WithContext("processor_type", processorType)
	return e
}

// WithProcessorName sets the processor name from configuration.
func (e *ProcessorError) WithProcessorName(name string) *ProcessorError {
	e.ProcessorName = name
	_ = e.EngineError.WithContext("processor_name", name)
	return e
}

// WithRecordIndex sets the index of the record that caused the error.
func (e *ProcessorError) WithRecordIndex(index int) *ProcessorError {
	e.RecordIndex = index
	_ = e.EngineError.WithContext("record_index", index)
	return e
}

// WithFieldName sets the field name that caused the error.
func (e *ProcessorError) WithFieldName(fieldName string) *ProcessorError {
	e.FieldName = fieldName
	_ = e.EngineError.WithContext("field_name", fieldName)
	return e
}

// WithContext adds context information to the error and returns the ProcessorError.
func (e *ProcessorError) WithContext(key string, value interface{}) *ProcessorError {
	_ = e.EngineError.WithContext(key, value)
	return e
}

// WithContextMap adds multiple context entries to the error and returns the ProcessorError.
func (e *ProcessorError) WithContextMap(ctx map[string]interface{}) *ProcessorError {
	_ = e.EngineError.WithContextMap(ctx)
	return e
}

// Common processor error constructors

// ErrProcessorValidation creates an error for validation failures.
func ErrProcessorValidation(processorName string, recordIndex int, fieldName string, err error) *ProcessorError {
	return NewProcessorError("validate", ErrorTypeValidation, err).
		WithProcessorName(processorName).
		WithRecordIndex(recordIndex).
		WithFieldName(fieldName)
}

// ErrProcessorFilter creates an error for filtering issues.
func ErrProcessorFilter(processorName string, recordIndex int, err error) *ProcessorError {
	return NewProcessorError("filter", ErrorTypePermanent, err).
		WithProcessorType("filter").
		WithProcessorName(processorName).
		WithRecordIndex(recordIndex)
}

// ErrProcessorTransform creates an error for transformation failures.
func ErrProcessorTransform(processorName string, recordIndex int, err error) *ProcessorError {
	return NewProcessorError("transform", ErrorTypePermanent, err).
		WithProcessorType("transformer").
		WithProcessorName(processorName).
		WithRecordIndex(recordIndex)
}

// ErrProcessorConfiguration creates an error for processor configuration issues.
func ErrProcessorConfiguration(processorName string, paramName string, err error) *ProcessorError {
	return NewProcessorError("configure", ErrorTypeConfiguration, err).
		WithProcessorName(processorName).
		WithContext("parameter", paramName)
}

// ErrProcessorMissingField creates an error when a required field is missing.
func ErrProcessorMissingField(processorName string, recordIndex int, fieldName string) *ProcessorError {
	return NewProcessorError("validate", ErrorTypeValidation,
		fmt.Errorf("required field '%s' is missing", fieldName)).
		WithProcessorName(processorName).
		WithRecordIndex(recordIndex).
		WithFieldName(fieldName)
}

// ErrProcessorInvalidType creates an error when a field has an invalid type.
func ErrProcessorInvalidType(processorName string, recordIndex int, fieldName string, expectedType string, actualType string) *ProcessorError {
	return NewProcessorError("validate", ErrorTypeValidation,
		fmt.Errorf("field '%s' expected type %s, got %s", fieldName, expectedType, actualType)).
		WithProcessorName(processorName).
		WithRecordIndex(recordIndex).
		WithFieldName(fieldName).
		WithContext("expected_type", expectedType).
		WithContext("actual_type", actualType)
}

// ErrProcessorChainInterrupted creates an error when the processor chain is interrupted.
func ErrProcessorChainInterrupted(processorName string, step int, err error) *ProcessorError {
	return NewProcessorError("process", ErrorTypePermanent, err).
		WithProcessorName(processorName).
		WithContext("chain_step", step)
}
