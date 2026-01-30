package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// Mock implementations for testing

// mockCloser is a simple closer that tracks if it was closed
type mockCloser struct {
	closed    bool
	shouldErr bool
	err       error
	mu        sync.Mutex
}

func (m *mockCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldErr {
		return m.err
	}

	m.closed = true
	return nil
}

func (m *mockCloser) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// mockContextCloser supports context-aware cleanup
type mockContextCloser struct {
	closed     bool
	closeDelay time.Duration
	shouldErr  bool
	err        error
	mu         sync.Mutex
}

func (m *mockContextCloser) CloseWithContext(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldErr {
		return m.err
	}

	// Simulate cleanup work
	if m.closeDelay > 0 {
		timer := time.NewTimer(m.closeDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
			// Cleanup completed
			m.closed = true
			return nil
		case <-ctx.Done():
			// Context canceled during cleanup
			return ctx.Err()
		}
	}

	m.closed = true
	return nil
}

func (m *mockContextCloser) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// TestMultiCloserAdd tests adding closers
func TestMultiCloserAdd(t *testing.T) {
	var mc MultiCloser

	c1 := &mockCloser{}
	c2 := &mockCloser{}

	mc.Add(c1)
	mc.Add(c2)

	if len(mc.closers) != 2 {
		t.Errorf("Expected 2 closers, got %d", len(mc.closers))
	}
}

// TestMultiCloserAddNil tests that nil closers are ignored
func TestMultiCloserAddNil(t *testing.T) {
	var mc MultiCloser

	mc.Add(nil)
	mc.Add(&mockCloser{})
	mc.Add(nil)

	if len(mc.closers) != 1 {
		t.Errorf("Expected 1 closer (nils ignored), got %d", len(mc.closers))
	}
}

// TestMultiCloserClose tests basic close functionality
func TestMultiCloserClose(t *testing.T) {
	var mc MultiCloser

	c1 := &mockCloser{}
	c2 := &mockCloser{}
	c3 := &mockCloser{}

	mc.Add(c1)
	mc.Add(c2)
	mc.Add(c3)

	err := mc.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Verify all were closed
	if !c1.isClosed() {
		t.Error("c1 was not closed")
	}
	if !c2.isClosed() {
		t.Error("c2 was not closed")
	}
	if !c3.isClosed() {
		t.Error("c3 was not closed")
	}
}

// orderTrackingCloser tracks the order in which closers are closed
type orderTrackingCloser struct {
	id         int
	closeOrder *[]int
	mu         *sync.Mutex
}

func (o *orderTrackingCloser) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	*o.closeOrder = append(*o.closeOrder, o.id)
	return nil
}

// TestMultiCloserCloseOrder tests LIFO close order
func TestMultiCloserCloseOrder(t *testing.T) {
	var mc MultiCloser
	var closeOrder []int
	var mu sync.Mutex

	// Create closers that record their close order
	for i := 1; i <= 3; i++ {
		closer := &orderTrackingCloser{
			id:         i,
			closeOrder: &closeOrder,
			mu:         &mu,
		}
		mc.Add(closer)
	}

	_ = mc.Close()

	// Verify LIFO order (3, 2, 1)
	expectedOrder := []int{3, 2, 1}
	if len(closeOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d closes, got %d", len(expectedOrder), len(closeOrder))
	}

	for i, expected := range expectedOrder {
		if closeOrder[i] != expected {
			t.Errorf("Close order[%d] = %d, expected %d", i, closeOrder[i], expected)
		}
	}
}

// TestMultiCloserCloseWithErrors tests error aggregation
func TestMultiCloserCloseWithErrors(t *testing.T) {
	var mc MultiCloser

	c1 := &mockCloser{shouldErr: true, err: errors.New("error1")}
	c2 := &mockCloser{} // Success
	c3 := &mockCloser{shouldErr: true, err: errors.New("error3")}

	mc.Add(c1)
	mc.Add(c2)
	mc.Add(c3)

	err := mc.Close()

	// Should get CloseErrors
	closeErrs, ok := err.(*CloseErrors)
	if !ok {
		t.Fatalf("Expected *CloseErrors, got %T", err)
	}

	// Should have 2 errors (c1 and c3)
	if len(closeErrs.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(closeErrs.Errors))
	}

	// Verify c2 was still closed despite errors
	if !c2.isClosed() {
		t.Error("c2 should have been closed despite other errors")
	}
}

// TestMultiCloserIdempotent tests multiple close calls
func TestMultiCloserIdempotent(t *testing.T) {
	var mc MultiCloser

	c := &mockCloser{}
	mc.Add(c)

	// First close
	err := mc.Close()
	if err != nil {
		t.Fatalf("First Close() failed: %v", err)
	}

	if !c.isClosed() {
		t.Error("Closer should be closed")
	}

	// Second close should be no-op
	err = mc.Close()
	if err != nil {
		t.Errorf("Second Close() should return nil, got: %v", err)
	}

	// Verify closers slice was cleared
	if len(mc.closers) != 0 {
		t.Errorf("Expected empty closers after first close, got %d", len(mc.closers))
	}
}

// TestMultiCloserEmpty tests closing with no closers
func TestMultiCloserEmpty(t *testing.T) {
	var mc MultiCloser

	err := mc.Close()
	if err != nil {
		t.Errorf("Empty MultiCloser.Close() should return nil, got: %v", err)
	}
}

// TestCloseErrorsError tests error message formatting
func TestCloseErrorsError(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		contains []string
	}{
		{
			name:     "no errors",
			errors:   []error{},
			contains: []string{"no errors"},
		},
		{
			name:     "single error",
			errors:   []error{errors.New("single error")},
			contains: []string{"single error"},
		},
		{
			name:     "multiple errors",
			errors:   []error{errors.New("error1"), errors.New("error2"), errors.New("error3")},
			contains: []string{"3 errors", "error1", "error2", "error3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := &CloseErrors{Errors: tt.errors}
			errMsg := ce.Error()

			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("Error message should contain %q, got: %s", substr, errMsg)
				}
			}
		})
	}
}

// TestCloseErrorsUnwrap tests error unwrapping
func TestCloseErrorsUnwrap(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errors.New("error2")

	ce := &CloseErrors{Errors: []error{err1, err2}}

	unwrapped := ce.Unwrap()

	if len(unwrapped) != 2 {
		t.Fatalf("Expected 2 unwrapped errors, got %d", len(unwrapped))
	}

	if unwrapped[0] != err1 {
		t.Error("First unwrapped error doesn't match")
	}

	if unwrapped[1] != err2 {
		t.Error("Second unwrapped error doesn't match")
	}
}

// TestCloseableInterface tests that mockCloser implements Closeable
func TestCloseableInterface(t *testing.T) {
	var _ Closeable = (*mockCloser)(nil)
}

// TestCloseableWithContextInterface tests interface implementation
func TestCloseableWithContextInterface(t *testing.T) {
	var _ CloseableWithContext = (*mockContextCloser)(nil)
}

// TestCloseableWithContextSuccess tests successful context-aware cleanup
func TestCloseableWithContextSuccess(t *testing.T) {
	closer := &mockContextCloser{}

	ctx := context.Background()
	err := closer.CloseWithContext(ctx)

	if err != nil {
		t.Errorf("CloseWithContext() failed: %v", err)
	}

	if !closer.isClosed() {
		t.Error("Closer should be closed")
	}
}

// TestCloseableWithContextCancellation tests context cancellation during cleanup
func TestCloseableWithContextCancellation(t *testing.T) {
	closer := &mockContextCloser{
		closeDelay: 100 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := closer.CloseWithContext(ctx)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	// Cleanup was interrupted, so should not be marked as closed
	if closer.isClosed() {
		t.Error("Closer should not be marked closed when context canceled")
	}
}

// TestCloseableWithContextTimeout tests cleanup timeout
func TestCloseableWithContextTimeout(t *testing.T) {
	closer := &mockContextCloser{
		closeDelay: 100 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := closer.CloseWithContext(ctx)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}

	if closer.isClosed() {
		t.Error("Closer should not be marked closed when deadline exceeded")
	}
}

// TestCloseableWithContextError tests cleanup with error
func TestCloseableWithContextError(t *testing.T) {
	expectedErr := errors.New("cleanup failed")
	closer := &mockContextCloser{
		shouldErr: true,
		err:       expectedErr,
	}

	ctx := context.Background()
	err := closer.CloseWithContext(ctx)

	if err != expectedErr {
		t.Errorf("Expected %v, got: %v", expectedErr, err)
	}
}

// TestMultiCloserConcurrentAdd tests that concurrent Add is NOT safe
// This test documents the expected behavior - Add should not be called concurrently
func TestMultiCloserConcurrentAddDocumentation(t *testing.T) {
	// This is a documentation test - MultiCloser.Add is NOT thread-safe by design
	// Users must ensure Add is only called during initialization
	// This test just verifies the documented behavior exists

	var mc MultiCloser
	mc.Add(&mockCloser{})

	// We document that concurrent Add is unsafe
	// In production, users should only call Add during setup, not concurrently
}

// TestMultiCloserConcurrentClose tests concurrent close safety
func TestMultiCloserConcurrentClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	var mc MultiCloser
	for i := 0; i < 10; i++ {
		mc.Add(&mockCloser{})
	}

	// Close concurrently from multiple goroutines
	var wg sync.WaitGroup
	errs := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- mc.Close()
		}()
	}

	wg.Wait()
	close(errs)

	// Only first close should do actual work, others return nil
	nilCount := 0
	for err := range errs {
		if err == nil {
			nilCount++
		}
	}

	// All calls should succeed (idempotent)
	if nilCount != 5 {
		t.Errorf("Expected all 5 concurrent closes to return nil, got %d nil returns", nilCount)
	}
}

// BenchmarkMultiCloserAdd benchmarks adding closers
func BenchmarkMultiCloserAdd(b *testing.B) {
	closer := &mockCloser{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var mc MultiCloser
		mc.Add(closer)
	}
}

// BenchmarkMultiCloserClose benchmarks closing
func BenchmarkMultiCloserClose(b *testing.B) {
	closers := make([]*mockCloser, 10)
	for i := range closers {
		closers[i] = &mockCloser{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var mc MultiCloser
		for _, c := range closers {
			mc.Add(c)
		}
		b.StartTimer()

		_ = mc.Close()
	}
}

// BenchmarkMultiCloserCloseWithErrors benchmarks error aggregation
func BenchmarkMultiCloserCloseWithErrors(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		var mc MultiCloser
		for j := 0; j < 10; j++ {
			c := &mockCloser{
				shouldErr: j%2 == 0,
				err:       fmt.Errorf("error%d", j),
			}
			mc.Add(c)
		}
		b.StartTimer()

		_ = mc.Close()
	}
}

// BenchmarkCloseErrorsError benchmarks error message formatting
func BenchmarkCloseErrorsError(b *testing.B) {
	errs := make([]error, 10)
	for i := range errs {
		errs[i] = fmt.Errorf("error%d", i)
	}

	ce := &CloseErrors{Errors: errs}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ce.Error()
	}
}
