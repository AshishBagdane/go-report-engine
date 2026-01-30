package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestErrorType tests the ErrorType string representation
func TestErrorType(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"unknown", ErrorTypeUnknown, "unknown"},
		{"validation", ErrorTypeValidation, "validation"},
		{"configuration", ErrorTypeConfiguration, "configuration"},
		{"transient", ErrorTypeTransient, "transient"},
		{"permanent", ErrorTypePermanent, "permanent"},
		{"resource", ErrorTypeResource, "resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// TestNewEngineError tests creation of EngineError
func TestNewEngineError(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	engineErr := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, baseErr)

	if engineErr.Component != ComponentProvider {
		t.Errorf("Component = %s, expected %s", engineErr.Component, ComponentProvider)
	}
	if engineErr.Operation != "fetch" {
		t.Errorf("Operation = %s, expected 'fetch'", engineErr.Operation)
	}
	if engineErr.Type != ErrorTypeTransient {
		t.Errorf("Type = %s, expected %s", engineErr.Type, ErrorTypeTransient)
	}
	if engineErr.Err != baseErr {
		t.Errorf("Err = %v, expected %v", engineErr.Err, baseErr)
	}
	if !engineErr.Retryable {
		t.Error("Retryable should be true for transient errors")
	}
	if engineErr.Context == nil {
		t.Error("Context should be initialized")
	}
	if engineErr.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

// TestEngineErrorError tests the Error() method
func TestEngineErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *EngineError
		contains []string
	}{
		{
			name: "basic error",
			err: NewEngineError(
				ComponentProvider,
				"fetch",
				ErrorTypeTransient,
				fmt.Errorf("connection timeout"),
			),
			contains: []string{"provider", "fetch", "connection timeout", "transient"},
		},
		{
			name: "error with context",
			err: NewEngineError(
				ComponentProcessor,
				"process",
				ErrorTypePermanent,
				fmt.Errorf("invalid data"),
			).WithContext("row", 42).WithContext("field", "email"),
			contains: []string{"processor", "process", "invalid data", "permanent", "row: 42", "field: email"},
		},
		{
			name: "error without type",
			err: &EngineError{
				Component: ComponentFormatter,
				Operation: "format",
				Type:      ErrorTypeUnknown,
				Err:       fmt.Errorf("format error"),
				Context:   make(map[string]interface{}),
			},
			contains: []string{"formatter", "format", "format error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errStr, substr) {
					t.Errorf("Error string should contain '%s', got: %s", substr, errStr)
				}
			}
		})
	}
}

// TestEngineErrorUnwrap tests the Unwrap() method
func TestEngineErrorUnwrap(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	engineErr := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, baseErr)

	unwrapped := engineErr.Unwrap()
	if unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, expected %v", unwrapped, baseErr)
	}
}

// TestEngineErrorTypeCheckers tests IsTransient, IsPermanent, etc.
func TestEngineErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name         string
		errorType    ErrorType
		isTransient  bool
		isPermanent  bool
		isValidation bool
		isConfig     bool
		isResource   bool
	}{
		{"transient", ErrorTypeTransient, true, false, false, false, false},
		{"permanent", ErrorTypePermanent, false, true, false, false, false},
		{"validation", ErrorTypeValidation, false, false, true, false, false},
		{"configuration", ErrorTypeConfiguration, false, false, false, true, false},
		{"resource", ErrorTypeResource, false, false, false, false, true},
		{"unknown", ErrorTypeUnknown, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewEngineError(ComponentProvider, "test", tt.errorType, fmt.Errorf("test"))

			if err.IsTransient() != tt.isTransient {
				t.Errorf("IsTransient() = %v, expected %v", err.IsTransient(), tt.isTransient)
			}
			if err.IsPermanent() != tt.isPermanent {
				t.Errorf("IsPermanent() = %v, expected %v", err.IsPermanent(), tt.isPermanent)
			}
			if err.IsValidation() != tt.isValidation {
				t.Errorf("IsValidation() = %v, expected %v", err.IsValidation(), tt.isValidation)
			}
			if err.IsConfiguration() != tt.isConfig {
				t.Errorf("IsConfiguration() = %v, expected %v", err.IsConfiguration(), tt.isConfig)
			}
			if err.IsResource() != tt.isResource {
				t.Errorf("IsResource() = %v, expected %v", err.IsResource(), tt.isResource)
			}
		})
	}
}

// TestWithContext tests adding context to errors
func TestWithContext(t *testing.T) {
	err := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, fmt.Errorf("error"))

	_ = err.WithContext("key1", "value1")
	_ = err.WithContext("key2", 42)
	_ = err.WithContext("key3", true)

	if len(err.Context) != 3 {
		t.Errorf("Context should have 3 entries, got %d", len(err.Context))
	}
	if err.Context["key1"] != "value1" {
		t.Errorf("Context[key1] = %v, expected 'value1'", err.Context["key1"])
	}
	if err.Context["key2"] != 42 {
		t.Errorf("Context[key2] = %v, expected 42", err.Context["key2"])
	}
	if err.Context["key3"] != true {
		t.Errorf("Context[key3] = %v, expected true", err.Context["key3"])
	}
}

// TestWithContextMap tests adding multiple context entries
func TestWithContextMap(t *testing.T) {
	err := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, fmt.Errorf("error"))

	contextMap := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	_ = err.WithContextMap(contextMap)

	if len(err.Context) != 3 {
		t.Errorf("Context should have 3 entries, got %d", len(err.Context))
	}
	for k, v := range contextMap {
		if err.Context[k] != v {
			t.Errorf("Context[%s] = %v, expected %v", k, err.Context[k], v)
		}
	}
}

// TestWrap tests the Wrap function
func TestWrap(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		component  Component
		operation  string
		shouldWrap bool
	}{
		{
			name:       "wrap standard error",
			err:        fmt.Errorf("standard error"),
			component:  ComponentProvider,
			operation:  "fetch",
			shouldWrap: true,
		},
		{
			name:       "wrap nil error",
			err:        nil,
			component:  ComponentProvider,
			operation:  "fetch",
			shouldWrap: false,
		},
		{
			name: "wrap engine error",
			err: NewEngineError(
				ComponentProcessor,
				"process",
				ErrorTypePermanent,
				fmt.Errorf("base"),
			),
			component:  ComponentEngine,
			operation:  "run",
			shouldWrap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := Wrap(tt.component, tt.operation, tt.err)

			if tt.shouldWrap {
				if wrapped == nil {
					t.Fatal("Wrap should return non-nil error")
				}

				engineErr, ok := wrapped.(*EngineError)
				if !ok {
					t.Fatal("Wrap should return EngineError")
				}

				if engineErr.Component != tt.component {
					t.Errorf("Component = %s, expected %s", engineErr.Component, tt.component)
				}
				if engineErr.Operation != tt.operation {
					t.Errorf("Operation = %s, expected %s", engineErr.Operation, tt.operation)
				}
			} else {
				if wrapped != nil {
					t.Error("Wrap should return nil for nil error")
				}
			}
		})
	}
}

// TestWrapWithType tests the WrapWithType function
func TestWrapWithType(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	wrapped := WrapWithType(ComponentProvider, "fetch", ErrorTypeTransient, baseErr)

	if wrapped == nil {
		t.Fatal("WrapWithType should return non-nil error")
	}

	engineErr, ok := wrapped.(*EngineError)
	if !ok {
		t.Fatal("WrapWithType should return EngineError")
	}

	if engineErr.Type != ErrorTypeTransient {
		t.Errorf("Type = %s, expected %s", engineErr.Type, ErrorTypeTransient)
	}
	if engineErr.Component != ComponentProvider {
		t.Errorf("Component = %s, expected %s", engineErr.Component, ComponentProvider)
	}

	// Test with nil error
	wrapped = WrapWithType(ComponentProvider, "fetch", ErrorTypeTransient, nil)
	if wrapped != nil {
		t.Error("WrapWithType should return nil for nil error")
	}
}

// TestIsEngineError tests the IsEngineError function
func TestIsEngineError(t *testing.T) {
	engineErr := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, fmt.Errorf("error"))
	standardErr := fmt.Errorf("standard error")

	if !IsEngineError(engineErr) {
		t.Error("IsEngineError should return true for EngineError")
	}
	if IsEngineError(standardErr) {
		t.Error("IsEngineError should return false for standard error")
	}
}

// TestGetErrorType tests the GetErrorType function
func TestGetErrorType(t *testing.T) {
	engineErr := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, fmt.Errorf("error"))
	standardErr := fmt.Errorf("standard error")

	if GetErrorType(engineErr) != ErrorTypeTransient {
		t.Errorf("GetErrorType(engineErr) = %s, expected %s", GetErrorType(engineErr), ErrorTypeTransient)
	}
	if GetErrorType(standardErr) != ErrorTypeUnknown {
		t.Errorf("GetErrorType(standardErr) = %s, expected %s", GetErrorType(standardErr), ErrorTypeUnknown)
	}
}

// TestIsRetryable tests the IsRetryable function
func TestIsRetryable(t *testing.T) {
	transientErr := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, fmt.Errorf("error"))
	permanentErr := NewEngineError(ComponentProvider, "fetch", ErrorTypePermanent, fmt.Errorf("error"))
	standardErr := fmt.Errorf("standard error")

	if !IsRetryable(transientErr) {
		t.Error("IsRetryable should return true for transient errors")
	}
	if IsRetryable(permanentErr) {
		t.Error("IsRetryable should return false for permanent errors")
	}
	if IsRetryable(standardErr) {
		t.Error("IsRetryable should return false for standard errors")
	}
}

// TestGetErrorChain tests the GetErrorChain function
func TestGetErrorChain(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	err1 := NewEngineError(ComponentProcessor, "process", ErrorTypePermanent, baseErr)
	err2 := NewEngineError(ComponentEngine, "run", ErrorTypeUnknown, err1)

	chain := GetErrorChain(err2)

	if len(chain) != 3 {
		t.Errorf("Chain should have 3 errors, got %d", len(chain))
	}
	if chain[0] != err2 {
		t.Error("First error in chain should be err2")
	}
	if chain[1] != err1 {
		t.Error("Second error in chain should be err1")
	}
	if chain[2] != baseErr {
		t.Error("Third error in chain should be baseErr")
	}
}

// TestGetRootCause tests the GetRootCause function
func TestGetRootCause(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	err1 := NewEngineError(ComponentProcessor, "process", ErrorTypePermanent, baseErr)
	err2 := NewEngineError(ComponentEngine, "run", ErrorTypeUnknown, err1)

	root := GetRootCause(err2)

	if root != baseErr {
		t.Errorf("Root cause should be baseErr, got %v", root)
	}

	// Test with single error
	root = GetRootCause(baseErr)
	if root != baseErr {
		t.Error("Root cause of single error should be itself")
	}
}

// TestErrorContext tests the ErrorContext builder
func TestErrorContext(t *testing.T) {
	ec := NewErrorContext(ComponentProvider, "fetch").
		WithType(ErrorTypeTransient).
		WithContext("retry_count", 3).
		WithContext("timeout", "30s")

	baseErr := fmt.Errorf("connection timeout")
	wrappedErr := ec.Wrap(baseErr)

	if wrappedErr == nil {
		t.Fatal("ErrorContext.Wrap should return non-nil error")
	}

	engineErr, ok := wrappedErr.(*EngineError)
	if !ok {
		t.Fatal("ErrorContext.Wrap should return EngineError")
	}

	if engineErr.Component != ComponentProvider {
		t.Errorf("Component = %s, expected %s", engineErr.Component, ComponentProvider)
	}
	if engineErr.Operation != "fetch" {
		t.Errorf("Operation = %s, expected 'fetch'", engineErr.Operation)
	}
	if engineErr.Type != ErrorTypeTransient {
		t.Errorf("Type = %s, expected %s", engineErr.Type, ErrorTypeTransient)
	}
	if len(engineErr.Context) != 2 {
		t.Errorf("Context should have 2 entries, got %d", len(engineErr.Context))
	}
	if engineErr.Context["retry_count"] != 3 {
		t.Error("Context should contain retry_count")
	}
}

// TestErrorContextNew tests creating new errors with ErrorContext
func TestErrorContextNew(t *testing.T) {
	ec := NewErrorContext(ComponentProvider, "fetch").
		WithType(ErrorTypePermanent)

	err := ec.New("invalid configuration")

	if err == nil {
		t.Fatal("ErrorContext.New should return non-nil error")
	}

	engineErr, ok := err.(*EngineError)
	if !ok {
		t.Fatal("ErrorContext.New should return EngineError")
	}

	if !strings.Contains(engineErr.Error(), "invalid configuration") {
		t.Error("Error should contain the message")
	}
}

// TestErrorContextWrapNil tests wrapping nil errors
func TestErrorContextWrapNil(t *testing.T) {
	ec := NewErrorContext(ComponentProvider, "fetch")
	err := ec.Wrap(nil)

	if err != nil {
		t.Error("ErrorContext.Wrap(nil) should return nil")
	}
}

// TestErrorChaining tests that errors.Is and errors.As work correctly
func TestErrorChaining(t *testing.T) {
	baseErr := fmt.Errorf("base error")
	wrappedErr := Wrap(ComponentProvider, "fetch", baseErr)

	// Test errors.Is
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find base error in chain")
	}

	// Test errors.As
	var engineErr *EngineError
	if !errors.As(wrappedErr, &engineErr) {
		t.Error("errors.As should find EngineError in chain")
	}
}

// BenchmarkNewEngineError benchmarks error creation
func BenchmarkNewEngineError(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, baseErr)
	}
}

// BenchmarkWrap benchmarks error wrapping
func BenchmarkWrap(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Wrap(ComponentProvider, "fetch", baseErr)
	}
}

// BenchmarkErrorString benchmarks error string generation
func BenchmarkErrorString(b *testing.B) {
	err := NewEngineError(ComponentProvider, "fetch", ErrorTypeTransient, fmt.Errorf("error")).
		WithContext("key1", "value1").
		WithContext("key2", 42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
