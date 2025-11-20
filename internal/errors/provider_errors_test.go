package errors

import (
	"fmt"
	"strings"
	"testing"
)

// TestNewProviderError tests creation of ProviderError
func TestNewProviderError(t *testing.T) {
	baseErr := fmt.Errorf("connection failed")
	provErr := NewProviderError("fetch", ErrorTypeTransient, baseErr)

	if provErr.Component != ComponentProvider {
		t.Errorf("Component = %s, expected %s", provErr.Component, ComponentProvider)
	}
	if provErr.Operation != "fetch" {
		t.Errorf("Operation = %s, expected 'fetch'", provErr.Operation)
	}
	if provErr.Type != ErrorTypeTransient {
		t.Errorf("Type = %s, expected %s", provErr.Type, ErrorTypeTransient)
	}
	if provErr.Err != baseErr {
		t.Errorf("Err = %v, expected %v", provErr.Err, baseErr)
	}
}

// TestProviderErrorWithProviderType tests setting provider type
func TestProviderErrorWithProviderType(t *testing.T) {
	provErr := NewProviderError("fetch", ErrorTypeTransient, fmt.Errorf("error")).
		WithProviderType("postgres")

	if provErr.ProviderType != "postgres" {
		t.Errorf("ProviderType = %s, expected 'postgres'", provErr.ProviderType)
	}
	if provErr.Context["provider_type"] != "postgres" {
		t.Error("Context should contain provider_type")
	}
}

// TestProviderErrorWithRecordCount tests setting record count
func TestProviderErrorWithRecordCount(t *testing.T) {
	provErr := NewProviderError("parse", ErrorTypePermanent, fmt.Errorf("error")).
		WithRecordCount(42)

	if provErr.RecordCount != 42 {
		t.Errorf("RecordCount = %d, expected 42", provErr.RecordCount)
	}
	if provErr.Context["record_count"] != 42 {
		t.Error("Context should contain record_count")
	}
}

// TestProviderErrorWithQuery tests setting query
func TestProviderErrorWithQuery(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectContext string
	}{
		{
			name:          "short query",
			query:         "SELECT * FROM users",
			expectContext: "query",
		},
		{
			name:          "long query",
			query:         strings.Repeat("SELECT * FROM users WHERE id = 1 AND ", 10),
			expectContext: "query_preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provErr := NewProviderError("query", ErrorTypePermanent, fmt.Errorf("error")).
				WithQuery(tt.query)

			if provErr.Query != tt.query {
				t.Error("Query should be set")
			}
			if _, ok := provErr.Context[tt.expectContext]; !ok {
				t.Errorf("Context should contain '%s'", tt.expectContext)
			}
		})
	}
}

// TestProviderErrorChaining tests method chaining
func TestProviderErrorChaining(t *testing.T) {
	provErr := NewProviderError("fetch", ErrorTypeTransient, fmt.Errorf("error")).
		WithProviderType("postgres").
		WithRecordCount(10).
		WithQuery("SELECT * FROM users").
		WithContext("additional", "info")

	if provErr.ProviderType != "postgres" {
		t.Error("ProviderType should be set")
	}
	if provErr.RecordCount != 10 {
		t.Error("RecordCount should be set")
	}
	if provErr.Query != "SELECT * FROM users" {
		t.Error("Query should be set")
	}
	if provErr.Context["additional"] != "info" {
		t.Error("Additional context should be set")
	}
}

// TestProviderErrorWithContextMap tests adding multiple context entries
func TestProviderErrorWithContextMap(t *testing.T) {
	contextMap := map[string]interface{}{
		"host": "localhost",
		"port": 5432,
		"db":   "testdb",
	}

	provErr := NewProviderError("connect", ErrorTypeTransient, fmt.Errorf("error")).
		WithContextMap(contextMap)

	for k, v := range contextMap {
		if provErr.Context[k] != v {
			t.Errorf("Context[%s] = %v, expected %v", k, provErr.Context[k], v)
		}
	}
}

// TestErrProviderConnection tests connection error constructor
func TestErrProviderConnection(t *testing.T) {
	baseErr := fmt.Errorf("connection refused")
	provErr := ErrProviderConnection("postgres", baseErr)

	if provErr.Operation != "connect" {
		t.Errorf("Operation = %s, expected 'connect'", provErr.Operation)
	}
	if provErr.Type != ErrorTypeTransient {
		t.Error("Connection errors should be transient")
	}
	if provErr.ProviderType != "postgres" {
		t.Error("ProviderType should be set")
	}
	if !provErr.Retryable {
		t.Error("Connection errors should be retryable")
	}
}

// TestErrProviderAuthentication tests authentication error constructor
func TestErrProviderAuthentication(t *testing.T) {
	baseErr := fmt.Errorf("invalid credentials")
	provErr := ErrProviderAuthentication("postgres", baseErr)

	if provErr.Operation != "authenticate" {
		t.Errorf("Operation = %s, expected 'authenticate'", provErr.Operation)
	}
	if provErr.Type != ErrorTypeConfiguration {
		t.Error("Authentication errors should be configuration errors")
	}
	if provErr.ProviderType != "postgres" {
		t.Error("ProviderType should be set")
	}
	if provErr.Retryable {
		t.Error("Authentication errors should not be retryable")
	}
}

// TestErrProviderQuery tests query error constructor
func TestErrProviderQuery(t *testing.T) {
	baseErr := fmt.Errorf("syntax error")
	query := "SELECT * FROM users WHERE invalid"
	provErr := ErrProviderQuery("postgres", query, baseErr)

	if provErr.Operation != "query" {
		t.Errorf("Operation = %s, expected 'query'", provErr.Operation)
	}
	if provErr.Type != ErrorTypePermanent {
		t.Error("Query errors should be permanent")
	}
	if provErr.ProviderType != "postgres" {
		t.Error("ProviderType should be set")
	}
	if provErr.Query != query {
		t.Error("Query should be set")
	}
	if provErr.Retryable {
		t.Error("Query errors should not be retryable")
	}
}

// TestErrProviderTimeout tests timeout error constructor
func TestErrProviderTimeout(t *testing.T) {
	baseErr := fmt.Errorf("operation timeout")
	provErr := ErrProviderTimeout("api", baseErr)

	if provErr.Operation != "fetch" {
		t.Errorf("Operation = %s, expected 'fetch'", provErr.Operation)
	}
	if provErr.Type != ErrorTypeTransient {
		t.Error("Timeout errors should be transient")
	}
	if provErr.ProviderType != "api" {
		t.Error("ProviderType should be set")
	}
	if !provErr.Retryable {
		t.Error("Timeout errors should be retryable")
	}
}

// TestErrProviderDataFormat tests data format error constructor
func TestErrProviderDataFormat(t *testing.T) {
	baseErr := fmt.Errorf("invalid JSON")
	provErr := ErrProviderDataFormat("api", 42, baseErr)

	if provErr.Operation != "parse" {
		t.Errorf("Operation = %s, expected 'parse'", provErr.Operation)
	}
	if provErr.Type != ErrorTypePermanent {
		t.Error("Data format errors should be permanent")
	}
	if provErr.ProviderType != "api" {
		t.Error("ProviderType should be set")
	}
	if provErr.RecordCount != 42 {
		t.Errorf("RecordCount = %d, expected 42", provErr.RecordCount)
	}
	if provErr.Retryable {
		t.Error("Data format errors should not be retryable")
	}
}

// TestErrProviderNotFound tests not found error constructor
func TestErrProviderNotFound(t *testing.T) {
	provErr := ErrProviderNotFound("csv", "/path/to/file.csv")

	if provErr.Operation != "fetch" {
		t.Errorf("Operation = %s, expected 'fetch'", provErr.Operation)
	}
	if provErr.Type != ErrorTypePermanent {
		t.Error("Not found errors should be permanent")
	}
	if provErr.ProviderType != "csv" {
		t.Error("ProviderType should be set")
	}
	if provErr.Context["resource"] != "/path/to/file.csv" {
		t.Error("Resource should be in context")
	}
	if !strings.Contains(provErr.Error(), "resource not found") {
		t.Error("Error should mention 'resource not found'")
	}
}

// TestErrProviderResourceExhausted tests resource exhaustion error constructor
func TestErrProviderResourceExhausted(t *testing.T) {
	baseErr := fmt.Errorf("connection pool exhausted")
	provErr := ErrProviderResourceExhausted("postgres", "connection_pool", baseErr)

	if provErr.Operation != "fetch" {
		t.Errorf("Operation = %s, expected 'fetch'", provErr.Operation)
	}
	if provErr.Type != ErrorTypeResource {
		t.Error("Resource exhaustion errors should be resource type")
	}
	if provErr.ProviderType != "postgres" {
		t.Error("ProviderType should be set")
	}
	if provErr.Context["exhausted_resource"] != "connection_pool" {
		t.Error("Exhausted resource should be in context")
	}
}

// TestProviderErrorUnwrap tests that ProviderError can be unwrapped
func TestProviderErrorUnwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	provErr := NewProviderError("fetch", ErrorTypeTransient, baseErr)

	unwrapped := provErr.Unwrap()
	if unwrapped != baseErr {
		t.Error("Unwrap should return the base error")
	}
}

// TestProviderErrorErrorString tests error string generation
func TestProviderErrorErrorString(t *testing.T) {
	provErr := NewProviderError("fetch", ErrorTypeTransient, fmt.Errorf("connection timeout")).
		WithProviderType("postgres").
		WithRecordCount(100)

	errStr := provErr.Error()

	// Should contain component and operation
	if !strings.Contains(errStr, "provider") {
		t.Error("Error string should contain component")
	}
	if !strings.Contains(errStr, "fetch") {
		t.Error("Error string should contain operation")
	}
	// Should contain the base error message
	if !strings.Contains(errStr, "connection timeout") {
		t.Error("Error string should contain base error message")
	}
	// Should contain type
	if !strings.Contains(errStr, "transient") {
		t.Error("Error string should contain error type")
	}
	// Should contain context
	if !strings.Contains(errStr, "provider_type") {
		t.Error("Error string should contain context")
	}
}

// TestProviderErrorTypeClassification tests error type classification
func TestProviderErrorTypeClassification(t *testing.T) {
	tests := []struct {
		name        string
		constructor func() *ProviderError
		isTransient bool
		isPermanent bool
		isRetryable bool
	}{
		{
			name:        "connection error",
			constructor: func() *ProviderError { return ErrProviderConnection("pg", fmt.Errorf("err")) },
			isTransient: true,
			isPermanent: false,
			isRetryable: true,
		},
		{
			name:        "authentication error",
			constructor: func() *ProviderError { return ErrProviderAuthentication("pg", fmt.Errorf("err")) },
			isTransient: false,
			isPermanent: false,
			isRetryable: false,
		},
		{
			name:        "query error",
			constructor: func() *ProviderError { return ErrProviderQuery("pg", "SELECT", fmt.Errorf("err")) },
			isTransient: false,
			isPermanent: true,
			isRetryable: false,
		},
		{
			name:        "timeout error",
			constructor: func() *ProviderError { return ErrProviderTimeout("api", fmt.Errorf("err")) },
			isTransient: true,
			isPermanent: false,
			isRetryable: true,
		},
		{
			name:        "data format error",
			constructor: func() *ProviderError { return ErrProviderDataFormat("csv", 1, fmt.Errorf("err")) },
			isTransient: false,
			isPermanent: true,
			isRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err.IsTransient() != tt.isTransient {
				t.Errorf("IsTransient() = %v, expected %v", err.IsTransient(), tt.isTransient)
			}
			if err.IsPermanent() != tt.isPermanent {
				t.Errorf("IsPermanent() = %v, expected %v", err.IsPermanent(), tt.isPermanent)
			}
			if err.Retryable != tt.isRetryable {
				t.Errorf("Retryable = %v, expected %v", err.Retryable, tt.isRetryable)
			}
		})
	}
}

// BenchmarkNewProviderError benchmarks provider error creation
func BenchmarkNewProviderError(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewProviderError("fetch", ErrorTypeTransient, baseErr)
	}
}

// BenchmarkProviderErrorWithChaining benchmarks error with chaining
func BenchmarkProviderErrorWithChaining(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewProviderError("fetch", ErrorTypeTransient, baseErr).
			WithProviderType("postgres").
			WithRecordCount(100).
			WithQuery("SELECT * FROM users")
	}
}

// BenchmarkProviderErrorString benchmarks error string generation
func BenchmarkProviderErrorString(b *testing.B) {
	err := NewProviderError("fetch", ErrorTypeTransient, fmt.Errorf("error")).
		WithProviderType("postgres").
		WithRecordCount(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
