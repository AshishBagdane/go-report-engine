// Package registry provides thread-safe registration and retrieval mechanisms
// for all pluggable components in the report engine.
package registry

import (
	"fmt"
	"sync"

	"github.com/AshishBagdane/go-report-engine/internal/output"
)

// OutputFactory is a function type that creates new instances of OutputStrategy.
// Factories are registered once (typically in init()) and called each time
// a new output instance is needed.
//
// Example:
//
//	func NewConsoleOutput() output.OutputStrategy {
//	    return &ConsoleOutput{}
//	}
type OutputFactory func() output.OutputStrategy

// outputRegistry holds the global registry of output factories.
// It uses a sync.RWMutex to ensure thread-safe access in concurrent environments.
var (
	outputRegistry   = make(map[string]OutputFactory)
	outputRegistryMu sync.RWMutex
)

// RegisterOutput registers an output factory with the given name.
// This function is typically called during package initialization (init functions)
// to register all available output implementations.
//
// The name parameter must be non-empty and should use lowercase with underscores
// (e.g., "console", "file", "s3_bucket", "slack_webhook"). Factory must not be nil.
//
// Registering the same name twice will overwrite the previous registration
// without error, allowing for intentional factory replacement in tests or
// custom implementations.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	func init() {
//	    registry.RegisterOutput("console", output.NewConsoleOutput)
//	    registry.RegisterOutput("file", output.NewFileOutput)
//	    registry.RegisterOutput("s3", output.NewS3Output)
//	}
//
// Parameters:
//   - name: Unique identifier for this output (must be non-empty)
//   - factory: Function that creates new output instances (must not be nil)
//
// Panics if name is empty or factory is nil, as these indicate programmer errors
// that should be caught during development/testing.
func RegisterOutput(name string, factory OutputFactory) {
	// Validate inputs - these are programmer errors and should fail fast
	if name == "" {
		panic("registry: output name cannot be empty")
	}
	if factory == nil {
		panic("registry: output factory cannot be nil")
	}

	outputRegistryMu.Lock()
	defer outputRegistryMu.Unlock()

	outputRegistry[name] = factory
}

// GetOutput retrieves an output factory by name and creates a new instance.
// This function is called at engine initialization time to construct the
// output component of the pipeline.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	output, err := registry.GetOutput("console")
//	if err != nil {
//	    log.Fatalf("Failed to get output: %v", err)
//	}
//	// Use output in pipeline
//
// Parameters:
//   - name: The name of the output to retrieve
//
// Returns:
//   - output.OutputStrategy: A new instance of the requested output
//   - error: ErrOutputNotFound if the name is not registered, or
//     ErrEmptyOutputName if name is empty
func GetOutput(name string) (output.OutputStrategy, error) {
	if name == "" {
		return nil, ErrEmptyOutputName
	}

	outputRegistryMu.RLock()
	factory, ok := outputRegistry[name]
	outputRegistryMu.RUnlock()

	if !ok {
		return nil, &ErrOutputNotFound{Name: name}
	}

	// Call factory outside the lock to minimize lock contention
	// and avoid potential deadlocks if factory itself accesses registries
	return factory(), nil
}

// ListOutputs returns a sorted list of all registered output names.
// This is useful for CLI help text, configuration validation, and debugging.
//
// Thread-safe: Yes. Returns a copy of the keys, so modifications to the
// returned slice do not affect the registry.
//
// Example:
//
//	outputs := registry.ListOutputs()
//	fmt.Printf("Available outputs: %v\n", outputs)
//
// Returns:
//   - []string: Alphabetically sorted list of registered output names
func ListOutputs() []string {
	outputRegistryMu.RLock()
	defer outputRegistryMu.RUnlock()

	names := make([]string, 0, len(outputRegistry))
	for name := range outputRegistry {
		names = append(names, name)
	}

	// Sort for consistent output
	sortStrings(names)
	return names
}

// IsOutputRegistered checks if an output with the given name exists.
// This is useful for configuration validation before attempting to build
// an engine.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	if !registry.IsOutputRegistered("console") {
//	    return fmt.Errorf("console output is required but not registered")
//	}
//
// Parameters:
//   - name: The output name to check
//
// Returns:
//   - bool: true if the output is registered, false otherwise
func IsOutputRegistered(name string) bool {
	if name == "" {
		return false
	}

	outputRegistryMu.RLock()
	_, ok := outputRegistry[name]
	outputRegistryMu.RUnlock()

	return ok
}

// UnregisterOutput removes an output from the registry.
// This is primarily useful for testing when you want to ensure a clean
// registry state between test cases.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// In test cleanup
//	registry.UnregisterOutput("test_output")
//
// Parameters:
//   - name: The output name to unregister
func UnregisterOutput(name string) {
	outputRegistryMu.Lock()
	defer outputRegistryMu.Unlock()

	delete(outputRegistry, name)
}

// ClearOutputs removes all outputs from the registry.
// This is primarily useful for testing when you need a completely clean state.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Warning: This will affect the global registry state. Use with caution,
// typically only in test code.
//
// Example:
//
//	func TestMain(m *testing.M) {
//	    // Clean slate for tests
//	    registry.ClearOutputs()
//	    // Register test outputs
//	    os.Exit(m.Run())
//	}
func ClearOutputs() {
	outputRegistryMu.Lock()
	defer outputRegistryMu.Unlock()

	outputRegistry = make(map[string]OutputFactory)
}

// OutputCount returns the number of registered outputs.
// This is useful for debugging and testing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	count := registry.OutputCount()
//	fmt.Printf("Total outputs registered: %d\n", count)
//
// Returns:
//   - int: The number of registered outputs
func OutputCount() int {
	outputRegistryMu.RLock()
	defer outputRegistryMu.RUnlock()

	return len(outputRegistry)
}

// ErrOutputNotFound is returned when a requested output is not registered.
type ErrOutputNotFound struct {
	Name string
}

func (e *ErrOutputNotFound) Error() string {
	return fmt.Sprintf("output not found: %s (available: %v)", e.Name, ListOutputs())
}

// ErrEmptyOutputName is returned when an empty string is provided as output name.
var ErrEmptyOutputName = fmt.Errorf("output name cannot be empty")
