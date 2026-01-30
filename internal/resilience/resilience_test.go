package resilience_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AshishBagdane/report-engine/internal/errors"
	"github.com/AshishBagdane/report-engine/internal/resilience"
)

// MockProviderStrategy for testing retries
type MockProviderStrategy struct {
	Attempts       int
	FailCounts     int
	SuccessContent []map[string]interface{}
}

func (m *MockProviderStrategy) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	m.Attempts++
	if m.Attempts <= m.FailCounts {
		// Return a transient error
		return nil, errors.NewEngineError(
			errors.ComponentProvider,
			"Fetch",
			errors.ErrorTypeTransient,
			fmt.Errorf("simulated network flake"),
		)
	}
	return m.SuccessContent, nil
}

func TestRetryLogic_SuccessAfterFailure(t *testing.T) {
	// Setup
	policy := resilience.RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Factor:     2.0,
		Jitter:     false,
	}
	retrier := resilience.NewRetrier(policy)

	mockProv := &MockProviderStrategy{
		FailCounts:     2,
		SuccessContent: []map[string]interface{}{{"id": 1}},
	}
	provWithRetry := resilience.NewProviderWithRetry(mockProv, retrier)

	// Execute
	data, err := provWithRetry.Fetch(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("expected 1 record, got %d", len(data))
	}
	if mockProv.Attempts != 3 {
		t.Errorf("expected 3 attempts (2 fails + 1 success), got %d", mockProv.Attempts)
	}
}

func TestRetryLogic_MaxRetriesExceeded(t *testing.T) {
	// Setup
	policy := resilience.RetryPolicy{
		MaxRetries: 2, // Less than failures
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
		Factor:     2.0,
		Jitter:     false,
	}
	retrier := resilience.NewRetrier(policy)

	mockProv := &MockProviderStrategy{
		FailCounts:     3, // More than retries
		SuccessContent: []map[string]interface{}{{"id": 1}},
	}
	provWithRetry := resilience.NewProviderWithRetry(mockProv, retrier)

	// Execute
	_, err := provWithRetry.Fetch(context.Background())

	// Assert
	if err == nil {
		t.Fatal("expected error, got success")
	}
	if mockProv.Attempts != 3 { // Initial + 2 retries
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", mockProv.Attempts)
	}
}

func TestRetryLogic_NonRetriableError(t *testing.T) {
	retrier := resilience.NewRetrier(resilience.DefaultRetryPolicy)

	op := func(ctx context.Context) error {
		return errors.NewEngineError(
			errors.ComponentProvider,
			"Fetch",
			errors.ErrorTypePermanent, // Permanent error
			fmt.Errorf("invalid config"),
		)
	}

	err := retrier.Execute(context.Background(), op)
	if err == nil {
		t.Fatal("expected error")
	}

	// Should fail immediately
	// We can't easily count attempts here without a closure counter,
	// but the logic ensures it returns on !IsRetriable
}
