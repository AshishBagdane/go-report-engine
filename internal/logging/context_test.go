package logging

import (
	"context"
	"testing"
)

// TestWithRequestID tests adding request ID to context
func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "req-abc-123"

	ctx = WithRequestID(ctx, requestID)

	retrieved := GetRequestID(ctx)
	if retrieved != requestID {
		t.Errorf("GetRequestID() = %q, expected %q", retrieved, requestID)
	}
}

// TestGetRequestIDEmpty tests retrieving from empty context
func TestGetRequestIDEmpty(t *testing.T) {
	ctx := context.Background()

	retrieved := GetRequestID(ctx)
	if retrieved != "" {
		t.Errorf("GetRequestID() = %q, expected empty string", retrieved)
	}
}

// TestWithCorrelationID tests adding correlation ID to context
func TestWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "corr-xyz-789"

	ctx = WithCorrelationID(ctx, correlationID)

	retrieved := GetCorrelationID(ctx)
	if retrieved != correlationID {
		t.Errorf("GetCorrelationID() = %q, expected %q", retrieved, correlationID)
	}
}

// TestGetCorrelationIDEmpty tests retrieving from empty context
func TestGetCorrelationIDEmpty(t *testing.T) {
	ctx := context.Background()

	retrieved := GetCorrelationID(ctx)
	if retrieved != "" {
		t.Errorf("GetCorrelationID() = %q, expected empty string", retrieved)
	}
}

// TestBothIDsInContext tests both IDs in same context
func TestBothIDsInContext(t *testing.T) {
	ctx := context.Background()
	requestID := "req-123"
	correlationID := "corr-456"

	ctx = WithRequestID(ctx, requestID)
	ctx = WithCorrelationID(ctx, correlationID)

	retrievedReq := GetRequestID(ctx)
	retrievedCorr := GetCorrelationID(ctx)

	if retrievedReq != requestID {
		t.Errorf("GetRequestID() = %q, expected %q", retrievedReq, requestID)
	}
	if retrievedCorr != correlationID {
		t.Errorf("GetCorrelationID() = %q, expected %q", retrievedCorr, correlationID)
	}
}

// TestContextChaining tests chaining context operations
func TestContextChaining(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithCorrelationID(ctx, "corr-1")

	// Create child context
	childCtx := WithRequestID(ctx, "req-2")

	// Parent should still have original request ID
	if GetRequestID(ctx) != "req-1" {
		t.Error("Parent context request ID should not change")
	}

	// Child should have new request ID
	if GetRequestID(childCtx) != "req-2" {
		t.Error("Child context should have new request ID")
	}

	// Both should have same correlation ID
	if GetCorrelationID(ctx) != "corr-1" {
		t.Error("Parent context should have correlation ID")
	}
	if GetCorrelationID(childCtx) != "corr-1" {
		t.Error("Child context should inherit correlation ID")
	}
}

// TestEmptyStringIDs tests that empty strings work correctly
func TestEmptyStringIDs(t *testing.T) {
	ctx := WithRequestID(context.Background(), "")
	ctx = WithCorrelationID(ctx, "")

	if GetRequestID(ctx) != "" {
		t.Error("Empty request ID should be retrievable as empty string")
	}
	if GetCorrelationID(ctx) != "" {
		t.Error("Empty correlation ID should be retrievable as empty string")
	}
}

// TestContextKeyCollision tests that keys don't collide
func TestContextKeyCollision(t *testing.T) {
	ctx := context.Background()

	// Set request ID
	ctx = WithRequestID(ctx, "request-value")

	// Set correlation ID
	ctx = WithCorrelationID(ctx, "correlation-value")

	// Both should be retrievable independently
	if GetRequestID(ctx) != "request-value" {
		t.Error("Request ID was affected by correlation ID")
	}
	if GetCorrelationID(ctx) != "correlation-value" {
		t.Error("Correlation ID was affected by request ID")
	}
}

// TestOverwriteRequestID tests overwriting request ID
func TestOverwriteRequestID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-1")
	ctx = WithRequestID(ctx, "req-2")

	retrieved := GetRequestID(ctx)
	if retrieved != "req-2" {
		t.Errorf("GetRequestID() = %q, expected %q (should be overwritten)", retrieved, "req-2")
	}
}

// TestOverwriteCorrelationID tests overwriting correlation ID
func TestOverwriteCorrelationID(t *testing.T) {
	ctx := WithCorrelationID(context.Background(), "corr-1")
	ctx = WithCorrelationID(ctx, "corr-2")

	retrieved := GetCorrelationID(ctx)
	if retrieved != "corr-2" {
		t.Errorf("GetCorrelationID() = %q, expected %q (should be overwritten)", retrieved, "corr-2")
	}
}

// BenchmarkWithRequestID benchmarks adding request ID to context
func BenchmarkWithRequestID(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithRequestID(ctx, "req-123")
	}
}

// BenchmarkGetRequestID benchmarks retrieving request ID from context
func BenchmarkGetRequestID(b *testing.B) {
	ctx := WithRequestID(context.Background(), "req-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetRequestID(ctx)
	}
}

// BenchmarkWithCorrelationID benchmarks adding correlation ID to context
func BenchmarkWithCorrelationID(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithCorrelationID(ctx, "corr-456")
	}
}

// BenchmarkGetCorrelationID benchmarks retrieving correlation ID from context
func BenchmarkGetCorrelationID(b *testing.B) {
	ctx := WithCorrelationID(context.Background(), "corr-456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCorrelationID(ctx)
	}
}
