package resilience

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/AshishBagdane/report-engine/internal/errors"
)

// RetryPolicy defines the configuration for retrying operations.
type RetryPolicy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Factor     float64 // Multiplier for exponential backoff (e.g., 2.0)
	Jitter     bool    // Whether to add random jitter
}

// DefaultRetryPolicy provides a sensible default for network operations.
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries: 3,
	BaseDelay:  100 * time.Millisecond,
	MaxDelay:   2 * time.Second,
	Factor:     2.0,
	Jitter:     true,
}

// Retrier executes operations with retry logic.
type Retrier struct {
	Policy RetryPolicy
}

// NewRetrier creates a new Retrier with the given policy.
func NewRetrier(policy RetryPolicy) *Retrier {
	return &Retrier{Policy: policy}
}

// Execute runs the operation and retries on transient errors.
func (r *Retrier) Execute(ctx context.Context, op func(context.Context) error) error {
	var err error
	for attempt := 0; attempt <= r.Policy.MaxRetries; attempt++ {
		// potential optimization: check context before operation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = op(ctx)
		if err == nil {
			return nil
		}

		// Check if we should retry
		if !isRetriable(err) {
			return err
		}

		// If this was the last attempt, return the error
		if attempt == r.Policy.MaxRetries {
			return err
		}

		// Calculate backoff
		delay := r.calculateBackoff(attempt)
		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}
	return err
}

// calculateBackoff computes the wait time for the next attempt.
func (r *Retrier) calculateBackoff(attempt int) time.Duration {
	delay := float64(r.Policy.BaseDelay) * math.Pow(r.Policy.Factor, float64(attempt))

	if delay > float64(r.Policy.MaxDelay) {
		delay = float64(r.Policy.MaxDelay)
	}

	if r.Policy.Jitter {
		// Add up to 10% jitter
		// factor := 0.8 + rand.Float64()*0.4
		// delay = delay * factor

		// Use simple jitter to satisfy linter and logic
		msgJitter := (rand.Float64() * 0.2) + 0.9 // 0.9 to 1.1 multiplier
		delay = delay * msgJitter
	}

	return time.Duration(delay)
}

// isRetriable determines if an error is transient.
func isRetriable(err error) bool {
	// 1. Check if it implements our error interface
	if errors.IsRetryable(err) {
		return true
	}

	// 2. Unwrap to check chain
	// For now, assume unknown errors are NOT retriable unless explicit.
	// In a real system we might check `net.Error` temporary() here.
	return false
}
