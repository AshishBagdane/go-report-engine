// Package errors provides formatter-specific error types and helpers.
package errors

import (
	"fmt"
)

// FormatterError represents errors that occur during data formatting.
type FormatterError struct {
	*EngineError
	FormatterType string
	OutputFormat  string
	RecordCount   int
}

// NewFormatterError creates a new formatter error.
func NewFormatterError(operation string, errorType ErrorType, err error) *FormatterError {
	return &FormatterError{
		EngineError: NewEngineError(ComponentFormatter, operation, errorType, err),
		RecordCount: -1, // -1 indicates not set
	}
}

// WithFormatterType sets the formatter type (e.g., "json", "csv", "xml").
func (e *FormatterError) WithFormatterType(formatterType string) *FormatterError {
	e.FormatterType = formatterType
	_ = e.EngineError.WithContext("formatter_type", formatterType)
	return e
}

// WithOutputFormat sets the desired output format.
func (e *FormatterError) WithOutputFormat(format string) *FormatterError {
	e.OutputFormat = format
	_ = e.EngineError.WithContext("output_format", format)
	return e
}

// WithRecordCount sets the number of records being formatted.
func (e *FormatterError) WithRecordCount(count int) *FormatterError {
	e.RecordCount = count
	_ = e.EngineError.WithContext("record_count", count)
	return e
}

// WithContext adds context information to the error and returns the FormatterError.
func (e *FormatterError) WithContext(key string, value interface{}) *FormatterError {
	_ = e.EngineError.WithContext(key, value)
	return e
}

// WithContextMap adds multiple context entries to the error and returns the FormatterError.
func (e *FormatterError) WithContextMap(ctx map[string]interface{}) *FormatterError {
	_ = e.EngineError.WithContextMap(ctx)
	return e
}

// Common formatter error constructors

// ErrFormatterEncoding creates an error for encoding failures.
func ErrFormatterEncoding(formatterType string, err error) *FormatterError {
	return NewFormatterError("encode", ErrorTypePermanent, err).
		WithFormatterType(formatterType)
}

// ErrFormatterInvalidData creates an error when data cannot be formatted.
func ErrFormatterInvalidData(formatterType string, recordCount int, err error) *FormatterError {
	return NewFormatterError("format", ErrorTypePermanent, err).
		WithFormatterType(formatterType).
		WithRecordCount(recordCount)
}

// ErrFormatterConfiguration creates an error for formatter configuration issues.
func ErrFormatterConfiguration(formatterType string, paramName string, err error) *FormatterError {
	return NewFormatterError("configure", ErrorTypeConfiguration, err).
		WithFormatterType(formatterType).
		WithContext("parameter", paramName)
}

// ErrFormatterUnsupportedType creates an error for unsupported data types.
func ErrFormatterUnsupportedType(formatterType string, dataType string) *FormatterError {
	return NewFormatterError("format", ErrorTypePermanent,
		fmt.Errorf("unsupported data type: %s", dataType)).
		WithFormatterType(formatterType).
		WithContext("data_type", dataType)
}

// ErrFormatterMemoryExhausted creates an error when memory is exhausted during formatting.
func ErrFormatterMemoryExhausted(formatterType string, recordCount int, err error) *FormatterError {
	return NewFormatterError("format", ErrorTypeResource, err).
		WithFormatterType(formatterType).
		WithRecordCount(recordCount)
}

// ErrFormatterSizeLimitExceeded creates an error when output size exceeds limits.
func ErrFormatterSizeLimitExceeded(formatterType string, size int64, limit int64) *FormatterError {
	return NewFormatterError("format", ErrorTypeResource,
		fmt.Errorf("output size %d bytes exceeds limit of %d bytes", size, limit)).
		WithFormatterType(formatterType).
		WithContext("output_size", size).
		WithContext("size_limit", limit)
}
