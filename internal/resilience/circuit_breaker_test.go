package resilience_test

import (
	"errors"
	"testing"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/resilience"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	threshold := uint(2)
	timeout := 100 * time.Millisecond
	cb := resilience.NewCircuitBreaker("test-cb", threshold, timeout)

	// 1. Initially Closed
	if cb.State() != resilience.StateClosed {
		t.Errorf("Expected Closed state, got %v", cb.State())
	}

	// 2. Fail up to threshold
	opFail := func() error { return errors.New("boom") }

	_ = cb.Execute(opFail) // 1 failure
	if cb.State() != resilience.StateClosed {
		t.Errorf("Expected Closed state after 1 failure, got %v", cb.State())
	}

	_ = cb.Execute(opFail) // 2 failures -> Open
	if cb.State() != resilience.StateOpen {
		t.Errorf("Expected Open state after 2 failures, got %v", cb.State())
	}

	// 3. Reject requests while Open
	err := cb.Execute(func() error { return nil })
	if err != resilience.ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	// 4. Wait for timeout -> HalfOpen
	time.Sleep(timeout + 10*time.Millisecond)

	// Next request should be allowed (HalfOpen probe)
	opSuccess := func() error { return nil }
	err = cb.Execute(opSuccess)
	if err != nil {
		t.Errorf("Expected success in HalfOpen, got %v", err)
	}

	// 5. Success in HalfOpen -> Closed
	if cb.State() != resilience.StateClosed {
		t.Errorf("Expected Closed state after recovery, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	threshold := uint(1)
	timeout := 50 * time.Millisecond
	cb := resilience.NewCircuitBreaker("test-cb", threshold, timeout)

	opFail := func() error { return errors.New("boom") }

	// Open the circuit
	_ = cb.Execute(opFail)
	if cb.State() != resilience.StateOpen {
		t.Fatalf("Failed to open circuit")
	}

	// Wait for timeout
	time.Sleep(timeout + 10*time.Millisecond)

	// Probe fails
	_ = cb.Execute(opFail)

	// Should revert to Open
	if cb.State() != resilience.StateOpen {
		t.Errorf("Expected Open state after failed probe, got %v", cb.State())
	}
}
