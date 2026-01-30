// Package errors provides provider-specific error types and helpers.
package errors

import (
	"fmt"
)

// ProviderError represents errors that occur during data fetching.
type ProviderError struct {
	*EngineError
	ProviderType string
	RecordCount  int
	Query        string
}

// NewProviderError creates a new provider error.
func NewProviderError(operation string, errorType ErrorType, err error) *ProviderError {
	return &ProviderError{
		EngineError: NewEngineError(ComponentProvider, operation, errorType, err),
	}
}

// WithProviderType sets the provider type (e.g., "postgres", "csv", "api").
func (e *ProviderError) WithProviderType(providerType string) *ProviderError {
	e.ProviderType = providerType
	_ = e.WithContext("provider_type", providerType)
	return e
}

// WithRecordCount sets the number of records processed before failure.
func (e *ProviderError) WithRecordCount(count int) *ProviderError {
	e.RecordCount = count
	_ = e.WithContext("record_count", count)
	return e
}

// WithQuery sets the query or request that caused the error.
func (e *ProviderError) WithQuery(query string) *ProviderError {
	e.Query = query
	// Don't add full query to context as it might be large or contain sensitive data
	// Instead, add a truncated version
	if len(query) > 100 {
		_ = e.EngineError.WithContext("query_preview", query[:100]+"...")
	} else {
		_ = e.EngineError.WithContext("query", query)
	}
	return e
}

// WithContext adds context information to the error and returns the ProviderError.
func (e *ProviderError) WithContext(key string, value interface{}) *ProviderError {
	_ = e.EngineError.WithContext(key, value)
	return e
}

// WithContextMap adds multiple context entries to the error and returns the ProviderError.
func (e *ProviderError) WithContextMap(ctx map[string]interface{}) *ProviderError {
	_ = e.EngineError.WithContextMap(ctx)
	return e
}

// Common provider error constructors

// ErrProviderConnection creates an error for connection failures.
func ErrProviderConnection(providerType string, err error) *ProviderError {
	return NewProviderError("connect", ErrorTypeTransient, err).
		WithProviderType(providerType)
}

// ErrProviderAuthentication creates an error for authentication failures.
func ErrProviderAuthentication(providerType string, err error) *ProviderError {
	return NewProviderError("authenticate", ErrorTypeConfiguration, err).
		WithProviderType(providerType)
}

// ErrProviderQuery creates an error for query execution failures.
func ErrProviderQuery(providerType string, query string, err error) *ProviderError {
	return NewProviderError("query", ErrorTypePermanent, err).
		WithProviderType(providerType).
		WithQuery(query)
}

// ErrProviderTimeout creates an error for timeout failures.
func ErrProviderTimeout(providerType string, err error) *ProviderError {
	return NewProviderError("fetch", ErrorTypeTransient, err).
		WithProviderType(providerType)
}

// ErrProviderDataFormat creates an error for data format issues.
func ErrProviderDataFormat(providerType string, recordNum int, err error) *ProviderError {
	return NewProviderError("parse", ErrorTypePermanent, err).
		WithProviderType(providerType).
		WithRecordCount(recordNum)
}

// ErrProviderNotFound creates an error when data source is not found.
func ErrProviderNotFound(providerType string, resource string) *ProviderError {
	return NewProviderError("fetch", ErrorTypePermanent, fmt.Errorf("resource not found: %s", resource)).
		WithProviderType(providerType).
		WithContext("resource", resource)
}

// ErrProviderResourceExhausted creates an error for resource exhaustion.
func ErrProviderResourceExhausted(providerType string, resource string, err error) *ProviderError {
	return NewProviderError("fetch", ErrorTypeResource, err).
		WithProviderType(providerType).
		WithContext("exhausted_resource", resource)
}
