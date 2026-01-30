// Package registry provides thread-safe registration and retrieval mechanisms
// for all pluggable components in the report engine.
package registry

import (
	"fmt"
	"sync"

	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// ProviderFactory is a function type that creates new instances of ProviderStrategy.
// Factories are registered once (typically in init()) and called each time
// a new provider instance is needed.
//
// Example:
//
//	func NewMockProvider() provider.ProviderStrategy {
//	    return &MockProvider{}
//	}
type ProviderFactory func() provider.ProviderStrategy

// providerRegistry holds the global registry of provider factories.
// It uses a sync.RWMutex to ensure thread-safe access in concurrent environments.
var (
	providerRegistry   = make(map[string]ProviderFactory)
	providerRegistryMu sync.RWMutex
)

// RegisterProvider registers a provider factory with the given name.
// This function is typically called during package initialization (init functions)
// to register all available provider implementations.
//
// The name parameter must be non-empty and should use lowercase with underscores
// (e.g., "mock", "sql_database", "csv_file", "rest_api"). Factory must not be nil.
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
//	    registry.RegisterProvider("mock", provider.NewMockProvider)
//	    registry.RegisterProvider("postgres", provider.NewPostgresProvider)
//	    registry.RegisterProvider("csv", provider.NewCSVProvider)
//	}
//
// Parameters:
//   - name: Unique identifier for this provider (must be non-empty)
//   - factory: Function that creates new provider instances (must not be nil)
//
// Panics if name is empty or factory is nil, as these indicate programmer errors
// that should be caught during development/testing.
func RegisterProvider(name string, factory ProviderFactory) {
	// Validate inputs - these are programmer errors and should fail fast
	if name == "" {
		panic("registry: provider name cannot be empty")
	}
	if factory == nil {
		panic("registry: provider factory cannot be nil")
	}

	providerRegistryMu.Lock()
	defer providerRegistryMu.Unlock()

	providerRegistry[name] = factory
}

// GetProvider retrieves a provider factory by name and creates a new instance.
// This function is called at engine initialization time to construct the
// data source component of the pipeline.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	provider, err := registry.GetProvider("mock")
//	if err != nil {
//	    log.Fatalf("Failed to get provider: %v", err)
//	}
//	// Use provider in pipeline
//
// Parameters:
//   - name: The name of the provider to retrieve
//
// Returns:
//   - provider.ProviderStrategy: A new instance of the requested provider
//   - error: ErrProviderNotFound if the name is not registered, or
//     ErrEmptyProviderName if name is empty
func GetProvider(name string) (provider.ProviderStrategy, error) {
	if name == "" {
		return nil, ErrEmptyProviderName
	}

	providerRegistryMu.RLock()
	factory, ok := providerRegistry[name]
	providerRegistryMu.RUnlock()

	if !ok {
		return nil, &ErrProviderNotFound{Name: name}
	}

	// Call factory outside the lock to minimize lock contention
	// and avoid potential deadlocks if factory itself accesses registries
	return factory(), nil
}

// ListProviders returns a sorted list of all registered provider names.
// This is useful for CLI help text, configuration validation, and debugging.
//
// Thread-safe: Yes. Returns a copy of the keys, so modifications to the
// returned slice do not affect the registry.
//
// Example:
//
//	providers := registry.ListProviders()
//	fmt.Printf("Available providers: %v\n", providers)
//
// Returns:
//   - []string: Alphabetically sorted list of registered provider names
func ListProviders() []string {
	providerRegistryMu.RLock()
	defer providerRegistryMu.RUnlock()

	names := make([]string, 0, len(providerRegistry))
	for name := range providerRegistry {
		names = append(names, name)
	}

	// Sort for consistent output
	sortStrings(names)
	return names
}

// IsProviderRegistered checks if a provider with the given name exists.
// This is useful for configuration validation before attempting to build
// an engine.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	if !registry.IsProviderRegistered("postgres") {
//	    return fmt.Errorf("postgres provider is required but not registered")
//	}
//
// Parameters:
//   - name: The provider name to check
//
// Returns:
//   - bool: true if the provider is registered, false otherwise
func IsProviderRegistered(name string) bool {
	if name == "" {
		return false
	}

	providerRegistryMu.RLock()
	_, ok := providerRegistry[name]
	providerRegistryMu.RUnlock()

	return ok
}

// UnregisterProvider removes a provider from the registry.
// This is primarily useful for testing when you want to ensure a clean
// registry state between test cases.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// In test cleanup
//	registry.UnregisterProvider("test_provider")
//
// Parameters:
//   - name: The provider name to unregister
func UnregisterProvider(name string) {
	providerRegistryMu.Lock()
	defer providerRegistryMu.Unlock()

	delete(providerRegistry, name)
}

// ClearProviders removes all providers from the registry.
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
//	    registry.ClearProviders()
//	    // Register test providers
//	    os.Exit(m.Run())
//	}
func ClearProviders() {
	providerRegistryMu.Lock()
	defer providerRegistryMu.Unlock()

	providerRegistry = make(map[string]ProviderFactory)
}

// ProviderCount returns the number of registered providers.
// This is useful for debugging and testing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	count := registry.ProviderCount()
//	fmt.Printf("Total providers registered: %d\n", count)
//
// Returns:
//   - int: The number of registered providers
func ProviderCount() int {
	providerRegistryMu.RLock()
	defer providerRegistryMu.RUnlock()

	return len(providerRegistry)
}

// ErrProviderNotFound is returned when a requested provider is not registered.
type ErrProviderNotFound struct {
	Name string
}

func (e *ErrProviderNotFound) Error() string {
	return fmt.Sprintf("provider not found: %s (available: %v)", e.Name, ListProviders())
}

// ErrEmptyProviderName is returned when an empty string is provided as provider name.
var ErrEmptyProviderName = fmt.Errorf("provider name cannot be empty")
