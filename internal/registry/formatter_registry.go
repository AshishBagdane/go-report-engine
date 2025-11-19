// Package registry provides thread-safe registration and retrieval mechanisms
// for all pluggable components in the report engine. This package implements
// the Registry pattern to enable dynamic component discovery and instantiation.
package registry

import (
	"fmt"
	"sync"

	"github.com/AshishBagdane/report-engine/internal/formatter"
)

// FormatterFactory is a function type that creates new instances of FormatStrategy.
// Factories are registered once (typically in init()) and called each time
// a new formatter instance is needed.
//
// Example:
//
//	func NewJSONFormatter() formatter.FormatStrategy {
//	    return &JSONFormatter{}
//	}
type FormatterFactory func() formatter.FormatStrategy

// formatterRegistry holds the global registry of formatter factories.
// It uses a sync.RWMutex to ensure thread-safe access in concurrent environments.
var (
	formatterRegistry   = make(map[string]FormatterFactory)
	formatterRegistryMu sync.RWMutex
)

// RegisterFormatter registers a formatter factory with the given name.
// This function is typically called during package initialization (init functions)
// to register all available formatters.
//
// The name parameter must be non-empty and should use lowercase with underscores
// (e.g., "json", "csv", "html_table"). Factory must not be nil.
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
//	    registry.RegisterFormatter("json", formatter.NewJSONFormatter)
//	    registry.RegisterFormatter("csv", formatter.NewCSVFormatter)
//	}
//
// Parameters:
//   - name: Unique identifier for this formatter (must be non-empty)
//   - factory: Function that creates new formatter instances (must not be nil)
//
// Panics if name is empty or factory is nil, as these indicate programmer errors
// that should be caught during development/testing.
func RegisterFormatter(name string, factory FormatterFactory) {
	// Validate inputs - these are programmer errors and should fail fast
	if name == "" {
		panic("registry: formatter name cannot be empty")
	}
	if factory == nil {
		panic("registry: formatter factory cannot be nil")
	}

	formatterRegistryMu.Lock()
	defer formatterRegistryMu.Unlock()

	formatterRegistry[name] = factory
}

// GetFormatter retrieves a formatter factory by name and creates a new instance.
// This function is called at engine initialization time to construct the
// formatting component of the pipeline.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	formatter, err := registry.GetFormatter("json")
//	if err != nil {
//	    log.Fatalf("Failed to get formatter: %v", err)
//	}
//	// Use formatter in pipeline
//
// Parameters:
//   - name: The name of the formatter to retrieve
//
// Returns:
//   - formatter.FormatStrategy: A new instance of the requested formatter
//   - error: ErrFormatterNotFound if the name is not registered, or
//     ErrEmptyFormatterName if name is empty
func GetFormatter(name string) (formatter.FormatStrategy, error) {
	if name == "" {
		return nil, ErrEmptyFormatterName
	}

	formatterRegistryMu.RLock()
	factory, ok := formatterRegistry[name]
	formatterRegistryMu.RUnlock()

	if !ok {
		return nil, &ErrFormatterNotFound{Name: name}
	}

	// Call factory outside the lock to minimize lock contention
	// and avoid potential deadlocks if factory itself accesses registries
	return factory(), nil
}

// ListFormatters returns a sorted list of all registered formatter names.
// This is useful for CLI help text, configuration validation, and debugging.
//
// Thread-safe: Yes. Returns a copy of the keys, so modifications to the
// returned slice do not affect the registry.
//
// Example:
//
//	formatters := registry.ListFormatters()
//	fmt.Printf("Available formatters: %v\n", formatters)
//
// Returns:
//   - []string: Alphabetically sorted list of registered formatter names
func ListFormatters() []string {
	formatterRegistryMu.RLock()
	defer formatterRegistryMu.RUnlock()

	names := make([]string, 0, len(formatterRegistry))
	for name := range formatterRegistry {
		names = append(names, name)
	}

	// Sort for consistent output
	sortStrings(names)
	return names
}

// IsFormatterRegistered checks if a formatter with the given name exists.
// This is useful for configuration validation before attempting to build
// an engine.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	if !registry.IsFormatterRegistered("json") {
//	    return fmt.Errorf("json formatter is required but not registered")
//	}
//
// Parameters:
//   - name: The formatter name to check
//
// Returns:
//   - bool: true if the formatter is registered, false otherwise
func IsFormatterRegistered(name string) bool {
	if name == "" {
		return false
	}

	formatterRegistryMu.RLock()
	_, ok := formatterRegistry[name]
	formatterRegistryMu.RUnlock()

	return ok
}

// UnregisterFormatter removes a formatter from the registry.
// This is primarily useful for testing when you want to ensure a clean
// registry state between test cases.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// In test cleanup
//	registry.UnregisterFormatter("test_formatter")
//
// Parameters:
//   - name: The formatter name to unregister
func UnregisterFormatter(name string) {
	formatterRegistryMu.Lock()
	defer formatterRegistryMu.Unlock()

	delete(formatterRegistry, name)
}

// ClearFormatters removes all formatters from the registry.
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
//	    registry.ClearFormatters()
//	    // Register test formatters
//	    os.Exit(m.Run())
//	}
func ClearFormatters() {
	formatterRegistryMu.Lock()
	defer formatterRegistryMu.Unlock()

	formatterRegistry = make(map[string]FormatterFactory)
}

// FormatterCount returns the number of registered formatters.
// This is useful for debugging and testing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	count := registry.FormatterCount()
//	fmt.Printf("Total formatters registered: %d\n", count)
//
// Returns:
//   - int: The number of registered formatters
func FormatterCount() int {
	formatterRegistryMu.RLock()
	defer formatterRegistryMu.RUnlock()

	return len(formatterRegistry)
}

// ErrFormatterNotFound is returned when a requested formatter is not registered.
type ErrFormatterNotFound struct {
	Name string
}

func (e *ErrFormatterNotFound) Error() string {
	return fmt.Sprintf("formatter not found: %s (available: %v)", e.Name, ListFormatters())
}

// ErrEmptyFormatterName is returned when an empty string is provided as formatter name.
var ErrEmptyFormatterName = fmt.Errorf("formatter name cannot be empty")

// sortStrings is a simple insertion sort for small slices.
// We avoid importing "sort" to keep dependencies minimal.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
