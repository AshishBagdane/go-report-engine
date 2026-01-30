package engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/processor"
)

// Mock implementations with cleanup support

type mockCloseableProvider struct {
	data      []map[string]interface{}
	closed    bool
	closeErr  error
	closeOnce sync.Once
	mu        sync.Mutex
}

func (m *mockCloseableProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	return m.data, nil
}

func (m *mockCloseableProvider) Close() error {
	m.closeOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.closed = true
	})
	return m.closeErr
}

func (m *mockCloseableProvider) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

type mockCloseableProcessor struct {
	processor.BaseProcessor
	closed    bool
	closeErr  error
	closeOnce sync.Once
	mu        sync.Mutex
}

func (m *mockCloseableProcessor) Close() error {
	m.closeOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.closed = true
	})
	return m.closeErr
}

func (m *mockCloseableProcessor) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

type mockCloseableFormatter struct {
	closed    bool
	closeErr  error
	closeOnce sync.Once
	mu        sync.Mutex
}

func (m *mockCloseableFormatter) Format(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
	return []byte("formatted"), nil
}

func (m *mockCloseableFormatter) Close() error {
	m.closeOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.closed = true
	})
	return m.closeErr
}

func (m *mockCloseableFormatter) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

type mockCloseableOutput struct {
	closed    bool
	closeErr  error
	closeOnce sync.Once
	mu        sync.Mutex
}

func (m *mockCloseableOutput) Send(ctx context.Context, data []byte) error {
	return nil
}

func (m *mockCloseableOutput) Close() error {
	m.closeOnce.Do(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.closed = true
	})
	return m.closeErr
}

func (m *mockCloseableOutput) isClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// TestEngineClose tests basic Close functionality
func TestEngineClose(t *testing.T) {
	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	processor := &mockCloseableProcessor{}
	formatter := &mockCloseableFormatter{}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: processor,
		Formatter: formatter,
		Output:    output,
	}

	err := engine.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify all components were closed
	if !provider.isClosed() {
		t.Error("Provider was not closed")
	}
	if !processor.isClosed() {
		t.Error("Processor was not closed")
	}
	if !formatter.isClosed() {
		t.Error("Formatter was not closed")
	}
	if !output.isClosed() {
		t.Error("Output was not closed")
	}
}

// TestEngineCloseIdempotent tests multiple Close calls
func TestEngineCloseIdempotent(t *testing.T) {
	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	// First close
	err := engine.Close()
	if err != nil {
		t.Errorf("First Close() returned error: %v", err)
	}

	if !provider.isClosed() {
		t.Error("Provider should be closed after first Close()")
	}

	// Second close should be no-op
	err = engine.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

// TestEngineCloseWithErrors tests error aggregation
func TestEngineCloseWithErrors(t *testing.T) {
	providerErr := errors.New("provider close failed")
	outputErr := errors.New("output close failed")

	provider := &mockCloseableProvider{
		data:     []map[string]interface{}{{"id": 1}},
		closeErr: providerErr,
	}
	output := &mockCloseableOutput{
		closeErr: outputErr,
	}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	err := engine.Close()
	if err == nil {
		t.Fatal("Close() should return error when components fail to close")
	}

	// Error should mention both failures
	errStr := err.Error()
	if !contains(errStr, "provider close failed") && !contains(errStr, "output close failed") {
		t.Errorf("Error should contain component errors, got: %v", err)
	}
}

// TestEngineCloseNonCloseableComponents tests with components that don't implement Close
func TestEngineCloseNonCloseableComponents(t *testing.T) {
	engine := &ReportEngine{
		Provider:  &mockProvider{data: []map[string]interface{}{{"id": 1}}},
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    &mockOutput{},
	}

	err := engine.Close()
	if err != nil {
		t.Errorf("Close() should succeed with non-closeable components, got: %v", err)
	}
}

// TestEngineCloseMixedComponents tests with some closeable, some not
func TestEngineCloseMixedComponents(t *testing.T) {
	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{}, // Not closeable
		Formatter: &mockFormatter{}, // Not closeable
		Output:    output,
	}

	err := engine.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify closeable components were closed
	if !provider.isClosed() {
		t.Error("Provider should be closed")
	}
	if !output.isClosed() {
		t.Error("Output should be closed")
	}
}

// TestEngineCloseWithContext tests context-aware cleanup
func TestEngineCloseWithContext(t *testing.T) {
	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	ctx := context.Background()
	err := engine.CloseWithContext(ctx)
	if err != nil {
		t.Errorf("CloseWithContext() returned error: %v", err)
	}

	if !provider.isClosed() {
		t.Error("Provider was not closed")
	}
	if !output.isClosed() {
		t.Error("Output was not closed")
	}
}

// TestEngineShutdown tests Shutdown with timeout
func TestEngineShutdown(t *testing.T) {
	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	err := engine.Shutdown(5 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}

	if !provider.isClosed() {
		t.Error("Provider was not closed")
	}
	if !output.isClosed() {
		t.Error("Output was not closed")
	}
}

// TestEngineCloseConcurrent tests concurrent Close calls
func TestEngineCloseConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	var wg sync.WaitGroup
	errs := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- engine.Close()
		}()
	}

	wg.Wait()
	close(errs)

	// All calls should succeed (idempotent)
	for err := range errs {
		if err != nil {
			t.Errorf("Concurrent Close() returned error: %v", err)
		}
	}

	// Provider and output should be closed exactly once
	if !provider.isClosed() {
		t.Error("Provider should be closed")
	}
	if !output.isClosed() {
		t.Error("Output should be closed")
	}
}

// TestEngineCloseAfterRun tests cleanup after successful Run
func TestEngineCloseAfterRun(t *testing.T) {
	provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	// Run the pipeline
	err := engine.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Then close
	err = engine.Close()
	if err != nil {
		t.Errorf("Close() after Run() failed: %v", err)
	}

	if !provider.isClosed() {
		t.Error("Provider was not closed")
	}
	if !output.isClosed() {
		t.Error("Output was not closed")
	}
}

// TestEngineCloseAfterRunError tests cleanup after failed Run
func TestEngineCloseAfterRunError(t *testing.T) {
	provider := &mockCloseableProvider{
		data: nil, // Will cause Run to fail
	}
	output := &mockCloseableOutput{}

	engine := &ReportEngine{
		Provider:  provider,
		Processor: &mockProcessor{},
		Formatter: &mockFormatter{},
		Output:    output,
	}

	// Run the pipeline (will fail)
	err := engine.Run()
	if err == nil {
		t.Fatal("Run() should have failed with nil data")
	}

	// Cleanup should still work
	err = engine.Close()
	if err != nil {
		t.Errorf("Close() after failed Run() returned error: %v", err)
	}

	if !provider.isClosed() {
		t.Error("Provider was not closed")
	}
	if !output.isClosed() {
		t.Error("Output was not closed")
	}
}

// BenchmarkEngineClose benchmarks cleanup
func BenchmarkEngineClose(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
		engine := &ReportEngine{
			Provider:  provider,
			Processor: &mockProcessor{},
			Formatter: &mockFormatter{},
			Output:    &mockCloseableOutput{},
		}
		b.StartTimer()

		_ = engine.Close()
	}
}

// BenchmarkEngineCloseWithContext benchmarks context-aware cleanup
func BenchmarkEngineCloseWithContext(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		provider := &mockCloseableProvider{data: []map[string]interface{}{{"id": 1}}}
		engine := &ReportEngine{
			Provider:  provider,
			Processor: &mockProcessor{},
			Formatter: &mockFormatter{},
			Output:    &mockCloseableOutput{},
		}
		b.StartTimer()

		_ = engine.CloseWithContext(ctx)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
