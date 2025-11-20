package errors

import (
	"fmt"
	"testing"
)

// --- Formatter Error Tests ---

// TestNewFormatterError tests creation of FormatterError
func TestNewFormatterError(t *testing.T) {
	baseErr := fmt.Errorf("encoding failed")
	fmtErr := NewFormatterError("encode", ErrorTypePermanent, baseErr)

	if fmtErr.Component != ComponentFormatter {
		t.Errorf("Component = %s, expected %s", fmtErr.Component, ComponentFormatter)
	}
	if fmtErr.Operation != "encode" {
		t.Errorf("Operation = %s, expected 'encode'", fmtErr.Operation)
	}
	if fmtErr.RecordCount != -1 {
		t.Errorf("RecordCount should be -1 by default, got %d", fmtErr.RecordCount)
	}
}

// TestFormatterErrorChaining tests method chaining
func TestFormatterErrorChaining(t *testing.T) {
	fmtErr := NewFormatterError("format", ErrorTypePermanent, fmt.Errorf("error")).
		WithFormatterType("json").
		WithOutputFormat("pretty").
		WithRecordCount(100).
		WithContext("additional", "info")

	if fmtErr.FormatterType != "json" {
		t.Error("FormatterType should be set")
	}
	if fmtErr.OutputFormat != "pretty" {
		t.Error("OutputFormat should be set")
	}
	if fmtErr.RecordCount != 100 {
		t.Error("RecordCount should be set")
	}
	if fmtErr.Context["additional"] != "info" {
		t.Error("Additional context should be set")
	}
}

// TestErrFormatterEncoding tests encoding error constructor
func TestErrFormatterEncoding(t *testing.T) {
	baseErr := fmt.Errorf("invalid character")
	fmtErr := ErrFormatterEncoding("json", baseErr)

	if fmtErr.Type != ErrorTypePermanent {
		t.Error("Encoding errors should be permanent")
	}
	if fmtErr.FormatterType != "json" {
		t.Error("FormatterType should be set")
	}
}

// TestErrFormatterInvalidData tests invalid data error constructor
func TestErrFormatterInvalidData(t *testing.T) {
	baseErr := fmt.Errorf("data cannot be formatted")
	fmtErr := ErrFormatterInvalidData("csv", 50, baseErr)

	if fmtErr.RecordCount != 50 {
		t.Error("RecordCount should be set")
	}
}

// TestErrFormatterConfiguration tests configuration error constructor
func TestErrFormatterConfiguration(t *testing.T) {
	baseErr := fmt.Errorf("invalid delimiter")
	fmtErr := ErrFormatterConfiguration("csv", "delimiter", baseErr)

	if fmtErr.Type != ErrorTypeConfiguration {
		t.Error("Should be configuration error type")
	}
	if fmtErr.Context["parameter"] != "delimiter" {
		t.Error("Parameter should be in context")
	}
}

// TestErrFormatterUnsupportedType tests unsupported type error constructor
func TestErrFormatterUnsupportedType(t *testing.T) {
	fmtErr := ErrFormatterUnsupportedType("json", "binary")

	if fmtErr.Context["data_type"] != "binary" {
		t.Error("Data type should be in context")
	}
}

// TestErrFormatterMemoryExhausted tests memory exhausted error constructor
func TestErrFormatterMemoryExhausted(t *testing.T) {
	baseErr := fmt.Errorf("out of memory")
	fmtErr := ErrFormatterMemoryExhausted("xml", 1000000, baseErr)

	if fmtErr.Type != ErrorTypeResource {
		t.Error("Should be resource error type")
	}
	if fmtErr.RecordCount != 1000000 {
		t.Error("RecordCount should be set")
	}
}

// TestErrFormatterSizeLimitExceeded tests size limit error constructor
func TestErrFormatterSizeLimitExceeded(t *testing.T) {
	fmtErr := ErrFormatterSizeLimitExceeded("json", 10*1024*1024, 5*1024*1024)

	if fmtErr.Type != ErrorTypeResource {
		t.Error("Should be resource error type")
	}
	if fmtErr.Context["output_size"] != int64(10*1024*1024) {
		t.Error("Output size should be in context")
	}
	if fmtErr.Context["size_limit"] != int64(5*1024*1024) {
		t.Error("Size limit should be in context")
	}
}

// --- Output Error Tests ---

// TestNewOutputError tests creation of OutputError
func TestNewOutputError(t *testing.T) {
	baseErr := fmt.Errorf("connection failed")
	outErr := NewOutputError("connect", ErrorTypeTransient, baseErr)

	if outErr.Component != ComponentOutput {
		t.Errorf("Component = %s, expected %s", outErr.Component, ComponentOutput)
	}
	if outErr.Operation != "connect" {
		t.Errorf("Operation = %s, expected 'connect'", outErr.Operation)
	}
	if outErr.DataSize != -1 {
		t.Errorf("DataSize should be -1 by default, got %d", outErr.DataSize)
	}
}

// TestOutputErrorChaining tests method chaining
func TestOutputErrorChaining(t *testing.T) {
	outErr := NewOutputError("send", ErrorTypeTransient, fmt.Errorf("error")).
		WithOutputType("file").
		WithDestination("/tmp/output.txt").
		WithDataSize(1024).
		WithContext("additional", "info")

	if outErr.OutputType != "file" {
		t.Error("OutputType should be set")
	}
	if outErr.Destination != "/tmp/output.txt" {
		t.Error("Destination should be set")
	}
	if outErr.DataSize != 1024 {
		t.Error("DataSize should be set")
	}
	if outErr.Context["additional"] != "info" {
		t.Error("Additional context should be set")
	}
}

// TestOutputErrorWithLongDestination tests destination truncation
func TestOutputErrorWithLongDestination(t *testing.T) {
	longDest := "https://example.com/very/long/path/" + string(make([]byte, 200))
	outErr := NewOutputError("send", ErrorTypeTransient, fmt.Errorf("error")).
		WithDestination(longDest)

	if outErr.Destination != longDest {
		t.Error("Full destination should be stored")
	}
	// Context should have truncated version
	if _, ok := outErr.Context["destination_preview"]; !ok {
		t.Error("Long destination should be truncated in context")
	}
}

// TestErrOutputConnection tests connection error constructor
func TestErrOutputConnection(t *testing.T) {
	baseErr := fmt.Errorf("connection refused")
	outErr := ErrOutputConnection("s3", "s3://bucket/file.txt", baseErr)

	if outErr.Type != ErrorTypeTransient {
		t.Error("Connection errors should be transient")
	}
	if outErr.OutputType != "s3" {
		t.Error("OutputType should be set")
	}
	if outErr.Retryable != true {
		t.Error("Connection errors should be retryable")
	}
}

// TestErrOutputAuthentication tests authentication error constructor
func TestErrOutputAuthentication(t *testing.T) {
	baseErr := fmt.Errorf("invalid credentials")
	outErr := ErrOutputAuthentication("s3", "s3://bucket/file.txt", baseErr)

	if outErr.Type != ErrorTypeConfiguration {
		t.Error("Authentication errors should be configuration errors")
	}
	if outErr.Retryable != false {
		t.Error("Authentication errors should not be retryable")
	}
}

// TestErrOutputWrite tests write error constructor
func TestErrOutputWrite(t *testing.T) {
	baseErr := fmt.Errorf("write failed")
	outErr := ErrOutputWrite("file", "/tmp/output.txt", baseErr)

	if outErr.Type != ErrorTypeTransient {
		t.Error("Write errors should be transient")
	}
	if outErr.Operation != "write" {
		t.Error("Operation should be 'write'")
	}
}

// TestErrOutputPermission tests permission error constructor
func TestErrOutputPermission(t *testing.T) {
	baseErr := fmt.Errorf("permission denied")
	outErr := ErrOutputPermission("file", "/tmp/output.txt", baseErr)

	if outErr.Type != ErrorTypeConfiguration {
		t.Error("Permission errors should be configuration errors")
	}
}

// TestErrOutputDiskFull tests disk full error constructor
func TestErrOutputDiskFull(t *testing.T) {
	outErr := ErrOutputDiskFull("file", "/tmp/output.txt", 1024*1024)

	if outErr.Type != ErrorTypeResource {
		t.Error("Disk full errors should be resource errors")
	}
	if outErr.DataSize != 1024*1024 {
		t.Error("DataSize should be set")
	}
}

// TestErrOutputTimeout tests timeout error constructor
func TestErrOutputTimeout(t *testing.T) {
	baseErr := fmt.Errorf("timeout")
	outErr := ErrOutputTimeout("api", "https://api.example.com", baseErr)

	if outErr.Type != ErrorTypeTransient {
		t.Error("Timeout errors should be transient")
	}
	if outErr.Retryable != true {
		t.Error("Timeout errors should be retryable")
	}
}

// TestErrOutputSizeLimitExceeded tests size limit error constructor
func TestErrOutputSizeLimitExceeded(t *testing.T) {
	outErr := ErrOutputSizeLimitExceeded("slack", 10*1024*1024, 5*1024*1024)

	if outErr.Type != ErrorTypeResource {
		t.Error("Should be resource error type")
	}
	if outErr.DataSize != 10*1024*1024 {
		t.Error("DataSize should be set")
	}
	if outErr.Context["size_limit"] != int64(5*1024*1024) {
		t.Error("Size limit should be in context")
	}
}

// TestErrOutputNotFound tests not found error constructor
func TestErrOutputNotFound(t *testing.T) {
	outErr := ErrOutputNotFound("file", "/tmp/nonexistent.txt")

	if outErr.Type != ErrorTypePermanent {
		t.Error("Not found errors should be permanent")
	}
	if outErr.Retryable != false {
		t.Error("Not found errors should not be retryable")
	}
}

// TestErrOutputConfiguration tests configuration error constructor
func TestErrOutputConfiguration(t *testing.T) {
	baseErr := fmt.Errorf("invalid bucket name")
	outErr := ErrOutputConfiguration("s3", "bucket", baseErr)

	if outErr.Type != ErrorTypeConfiguration {
		t.Error("Should be configuration error type")
	}
	if outErr.Context["parameter"] != "bucket" {
		t.Error("Parameter should be in context")
	}
}

// TestErrOutputRateLimitExceeded tests rate limit error constructor
func TestErrOutputRateLimitExceeded(t *testing.T) {
	outErr := ErrOutputRateLimitExceeded("api", "https://api.example.com", "100 req/min")

	if outErr.Type != ErrorTypeTransient {
		t.Error("Rate limit errors should be transient")
	}
	if outErr.Context["rate_limit"] != "100 req/min" {
		t.Error("Rate limit should be in context")
	}
	if outErr.Retryable != true {
		t.Error("Rate limit errors should be retryable")
	}
}

// --- Benchmarks ---

// BenchmarkNewFormatterError benchmarks formatter error creation
func BenchmarkNewFormatterError(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewFormatterError("format", ErrorTypePermanent, baseErr)
	}
}

// BenchmarkNewOutputError benchmarks output error creation
func BenchmarkNewOutputError(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewOutputError("send", ErrorTypeTransient, baseErr)
	}
}

// BenchmarkOutputErrorWithChaining benchmarks error with chaining
func BenchmarkOutputErrorWithChaining(b *testing.B) {
	baseErr := fmt.Errorf("base error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewOutputError("send", ErrorTypeTransient, baseErr).
			WithOutputType("file").
			WithDestination("/tmp/output.txt").
			WithDataSize(1024)
	}
}
