package resilience

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

var (
	// ErrCircuitOpen is returned when the circuit breaker rejects a request.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the Circuit Breaker pattern.
type CircuitBreaker struct {
	name             string
	failureThreshold uint
	resetTimeout     time.Duration

	mu          sync.Mutex
	state       CircuitState
	failures    uint
	lastFailure time.Time
}

// NewCircuitBreaker creates a new CircuitBreaker.
func NewCircuitBreaker(name string, threshold uint, timeout time.Duration) *CircuitBreaker {
	if threshold == 0 {
		threshold = 5 // Default
	}
	if timeout == 0 {
		timeout = 60 * time.Second // Default
	}
	return &CircuitBreaker{
		name:             name,
		failureThreshold: threshold,
		resetTimeout:     timeout,
		state:            StateClosed,
	}
}

// Execute runs the given operation if the circuit is allowed.
func (cb *CircuitBreaker) Execute(op func() error) error {
	if !cb.Allow() {
		return ErrCircuitOpen
	}

	err := op()

	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// Allow checks if a request should be allowed to proceed.
// It handles the transition from Open to HalfOpen based on timeout.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateClosed {
		return true
	}

	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	}

	// StateHalfOpen
	// We allow the request to probe the service.
	return true
}

// RecordFailure records a failure and updates the state.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		return
	}

	if cb.state == StateClosed && cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// RecordSuccess records a success and updates the state.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateClosed
		cb.failures = 0
	case StateClosed:
		cb.failures = 0
	}
}

// State returns the current state (thread-safe).
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
