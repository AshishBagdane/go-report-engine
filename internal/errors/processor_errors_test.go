package errors

import (
	"fmt"
	"strings"
	"testing"
)

// TestNewProcessorError tests creation of ProcessorError
func TestNewProcessorError(t *testing.T) {
	baseErr := fmt.Errorf("validation failed")
	procErr := NewProcessorError("validate", ErrorTypeValidation, baseErr)

	if procErr.Component != ComponentProcessor {
		t.Errorf("Component = %s, expected %s", procErr.Component, ComponentProcessor)
	}
	if procErr.Operation != "validate" {
		t.Errorf("Operation = %s, expected 'validate'", procErr.Operation)
	}
	if procErr.RecordIndex != -1 {
		t.Errorf("RecordIndex should be -1 by default, got %d", procErr.RecordIndex)
	}
}

// TestProcessorErrorChaining tests method chaining
func TestProcessorErrorChaining(t *testing.T) {
	procErr := NewProcessorError("validate", ErrorTypeValidation, fmt.Errorf("error")).
		WithProcessorType("filter").
		WithProcessorName("min_score").
		WithRecordIndex(42).
		WithFieldName("email").
		WithContext("additional", "info")

	if procErr.ProcessorType != "filter" {
		t.Error("ProcessorType should be set")
	}
	if procErr.ProcessorName != "min_score" {
		t.Error("ProcessorName should be set")
	}
	if procErr.RecordIndex != 42 {
		t.Error("RecordIndex should be set")
	}
	if procErr.FieldName != "email" {
		t.Error("FieldName should be set")
	}
	if procErr.Context["additional"] != "info" {
		t.Error("Additional context should be set")
	}
}

// TestErrProcessorValidation tests validation error constructor
func TestErrProcessorValidation(t *testing.T) {
	baseErr := fmt.Errorf("invalid email format")
	procErr := ErrProcessorValidation("email_validator", 10, "email", baseErr)

	if procErr.Type != ErrorTypeValidation {
		t.Error("Should be validation error type")
	}
	if procErr.ProcessorName != "email_validator" {
		t.Error("ProcessorName should be set")
	}
	if procErr.RecordIndex != 10 {
		t.Error("RecordIndex should be set")
	}
	if procErr.FieldName != "email" {
		t.Error("FieldName should be set")
	}
}

// TestErrProcessorFilter tests filter error constructor
func TestErrProcessorFilter(t *testing.T) {
	baseErr := fmt.Errorf("filter failed")
	procErr := ErrProcessorFilter("score_filter", 5, baseErr)

	if procErr.Type != ErrorTypePermanent {
		t.Error("Filter errors should be permanent")
	}
	if procErr.ProcessorType != "filter" {
		t.Error("ProcessorType should be 'filter'")
	}
	if procErr.ProcessorName != "score_filter" {
		t.Error("ProcessorName should be set")
	}
}

// TestErrProcessorTransform tests transform error constructor
func TestErrProcessorTransform(t *testing.T) {
	baseErr := fmt.Errorf("transform failed")
	procErr := ErrProcessorTransform("uppercase", 3, baseErr)

	if procErr.Type != ErrorTypePermanent {
		t.Error("Transform errors should be permanent")
	}
	if procErr.ProcessorType != "transformer" {
		t.Error("ProcessorType should be 'transformer'")
	}
}

// TestErrProcessorConfiguration tests configuration error constructor
func TestErrProcessorConfiguration(t *testing.T) {
	baseErr := fmt.Errorf("invalid parameter value")
	procErr := ErrProcessorConfiguration("validator", "min_value", baseErr)

	if procErr.Type != ErrorTypeConfiguration {
		t.Error("Should be configuration error type")
	}
	if procErr.Context["parameter"] != "min_value" {
		t.Error("Parameter should be in context")
	}
}

// TestErrProcessorMissingField tests missing field error constructor
func TestErrProcessorMissingField(t *testing.T) {
	procErr := ErrProcessorMissingField("required_fields", 7, "user_id")

	if procErr.Type != ErrorTypeValidation {
		t.Error("Should be validation error type")
	}
	if procErr.FieldName != "user_id" {
		t.Error("FieldName should be set")
	}
	if !strings.Contains(procErr.Error(), "user_id") {
		t.Error("Error should mention the field name")
	}
}

// TestErrProcessorInvalidType tests invalid type error constructor
func TestErrProcessorInvalidType(t *testing.T) {
	procErr := ErrProcessorInvalidType("type_checker", 12, "age", "int", "string")

	if procErr.Type != ErrorTypeValidation {
		t.Error("Should be validation error type")
	}
	if procErr.Context["expected_type"] != "int" {
		t.Error("Expected type should be in context")
	}
	if procErr.Context["actual_type"] != "string" {
		t.Error("Actual type should be in context")
	}
}

// TestErrProcessorChainInterrupted tests chain interrupted error constructor
func TestErrProcessorChainInterrupted(t *testing.T) {
	baseErr := fmt.Errorf("processor failed")
	procErr := ErrProcessorChainInterrupted("validator", 3, baseErr)

	if procErr.Context["chain_step"] != 3 {
		t.Error("Chain step should be in context")
	}
}

// BenchmarkNewProcessorError benchmarks processor error creation
func BenchmarkNewProcessorError(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewProcessorError("validate", ErrorTypeValidation, baseErr)
	}
}

// BenchmarkProcessorErrorWithChaining benchmarks error with chaining
func BenchmarkProcessorErrorWithChaining(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewProcessorError("validate", ErrorTypeValidation, baseErr).
			WithProcessorType("filter").
			WithProcessorName("min_score").
			WithRecordIndex(42)
	}
}
