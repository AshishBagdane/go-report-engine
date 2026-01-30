// Package registry provides thread-safe registration and retrieval mechanisms
// for all pluggable components in the report engine.
package registry

import (
	"fmt"
	"sync"

	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/pkg/api"
)

// ProcessorFactory is a function type that creates new instances of ProcessorHandler.
// Factories are registered once (typically in init()) and called each time
// a new processor instance is needed for a pipeline.
//
// Example:
//
//	func NewSanitizeProcessor() processor.ProcessorHandler {
//	    return &SanitizeProcessor{}
//	}
type ProcessorFactory func() processor.ProcessorHandler

// processorRegistry holds the global registry of processor factories.
// It uses a sync.RWMutex to ensure thread-safe access in concurrent environments.
var (
	processorRegistry   = make(map[string]ProcessorFactory)
	processorRegistryMu sync.RWMutex
)

// RegisterProcessor registers a processor factory with the given name.
// This is the low-level registration function that accepts any ProcessorHandler factory.
//
// For most use cases, prefer the type-safe helpers:
//   - RegisterFilter() for FilterStrategy implementations
//   - RegisterValidator() for ValidatorStrategy implementations
//   - RegisterTransformer() for TransformerStrategy implementations
//
// The name parameter must be non-empty and should use lowercase with underscores
// (e.g., "sanitize", "aggregate", "deduplicate"). Factory must not be nil.
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
//	    registry.RegisterProcessor("custom", func() processor.ProcessorHandler {
//	        return &CustomProcessor{}
//	    })
//	}
//
// Parameters:
//   - name: Unique identifier for this processor (must be non-empty)
//   - factory: Function that creates new processor instances (must not be nil)
//
// Panics if name is empty or factory is nil, as these indicate programmer errors
// that should be caught during development/testing.
func RegisterProcessor(name string, factory ProcessorFactory) {
	// Validate inputs - these are programmer errors and should fail fast
	if name == "" {
		panic("registry: processor name cannot be empty")
	}
	if factory == nil {
		panic("registry: processor factory cannot be nil")
	}

	processorRegistryMu.Lock()
	defer processorRegistryMu.Unlock()

	processorRegistry[name] = factory
}

// GetProcessor retrieves a processor factory by name and creates a new instance.
// This function is called during processor chain construction to build the
// processing pipeline.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	proc, err := registry.GetProcessor("sanitize")
//	if err != nil {
//	    log.Fatalf("Failed to get processor: %v", err)
//	}
//	// Use processor in chain
//
// Parameters:
//   - name: The name of the processor to retrieve
//
// Returns:
//   - processor.ProcessorHandler: A new instance of the requested processor
//   - error: ErrProcessorNotFound if the name is not registered, or
//     ErrEmptyProcessorName if name is empty
func GetProcessor(name string) (processor.ProcessorHandler, error) {
	if name == "" {
		return nil, ErrEmptyProcessorName
	}

	processorRegistryMu.RLock()
	factory, ok := processorRegistry[name]
	processorRegistryMu.RUnlock()

	if !ok {
		return nil, &ErrProcessorNotFound{Name: name}
	}

	// Call factory outside the lock to minimize lock contention
	// and avoid potential deadlocks if factory itself accesses registries
	return factory(), nil
}

// ListProcessors returns a sorted list of all registered processor names.
// This is useful for CLI help text, configuration validation, and debugging.
//
// Thread-safe: Yes. Returns a copy of the keys, so modifications to the
// returned slice do not affect the registry.
//
// Example:
//
//	processors := registry.ListProcessors()
//	fmt.Printf("Available processors: %v\n", processors)
//
// Returns:
//   - []string: Alphabetically sorted list of registered processor names
func ListProcessors() []string {
	processorRegistryMu.RLock()
	defer processorRegistryMu.RUnlock()

	names := make([]string, 0, len(processorRegistry))
	for name := range processorRegistry {
		names = append(names, name)
	}

	// Sort for consistent output
	sortStrings(names)
	return names
}

// IsProcessorRegistered checks if a processor with the given name exists.
// This is useful for configuration validation before attempting to build
// a processor chain.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	if !registry.IsProcessorRegistered("sanitize") {
//	    return fmt.Errorf("sanitize processor is required but not registered")
//	}
//
// Parameters:
//   - name: The processor name to check
//
// Returns:
//   - bool: true if the processor is registered, false otherwise
func IsProcessorRegistered(name string) bool {
	if name == "" {
		return false
	}

	processorRegistryMu.RLock()
	_, ok := processorRegistry[name]
	processorRegistryMu.RUnlock()

	return ok
}

// UnregisterProcessor removes a processor from the registry.
// This is primarily useful for testing when you want to ensure a clean
// registry state between test cases.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// In test cleanup
//	registry.UnregisterProcessor("test_processor")
//
// Parameters:
//   - name: The processor name to unregister
func UnregisterProcessor(name string) {
	processorRegistryMu.Lock()
	defer processorRegistryMu.Unlock()

	delete(processorRegistry, name)
}

// ClearProcessors removes all processors from the registry.
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
//	    registry.ClearProcessors()
//	    // Register test processors
//	    os.Exit(m.Run())
//	}
func ClearProcessors() {
	processorRegistryMu.Lock()
	defer processorRegistryMu.Unlock()

	processorRegistry = make(map[string]ProcessorFactory)
}

// ProcessorCount returns the number of registered processors.
// This is useful for debugging and testing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	count := registry.ProcessorCount()
//	fmt.Printf("Total processors registered: %d\n", count)
//
// Returns:
//   - int: The number of registered processors
func ProcessorCount() int {
	processorRegistryMu.RLock()
	defer processorRegistryMu.RUnlock()

	return len(processorRegistry)
}

// --- Type-Safe Registration Helpers ---

// RegisterFilter registers a FilterStrategy by automatically wrapping it
// in a FilterWrapper. This is the recommended way to register filters as it
// provides type safety and reduces boilerplate.
//
// The filter will be automatically wrapped to implement ProcessorHandler
// and participate in the Chain of Responsibility pattern.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	type MinScoreFilter struct {
//	    MinScore int
//	}
//
//	func (f *MinScoreFilter) Keep(row map[string]interface{}) bool {
//	    if score, ok := row["score"].(int); ok {
//	        return score >= f.MinScore
//	    }
//	    return false
//	}
//
//	func init() {
//	    registry.RegisterFilter("min_score", &MinScoreFilter{MinScore: 80})
//	}
//
// Parameters:
//   - name: Unique identifier for this filter (must be non-empty)
//   - strategy: FilterStrategy implementation (must not be nil)
//
// Panics if name is empty or strategy is nil.
func RegisterFilter(name string, strategy api.FilterStrategy) {
	if strategy == nil {
		panic("registry: filter strategy cannot be nil")
	}

	RegisterProcessor(name, func() processor.ProcessorHandler {
		return processor.NewFilterWrapper(strategy)
	})
}

// RegisterValidator registers a ValidatorStrategy by automatically wrapping it
// in a ValidatorWrapper. This is the recommended way to register validators as it
// provides type safety and reduces boilerplate.
//
// The validator will be automatically wrapped to implement ProcessorHandler
// and participate in the Chain of Responsibility pattern.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	type EmailValidator struct{}
//
//	func (v *EmailValidator) Validate(row map[string]interface{}) error {
//	    email, ok := row["email"].(string)
//	    if !ok || !strings.Contains(email, "@") {
//	        return fmt.Errorf("invalid email format")
//	    }
//	    return nil
//	}
//
//	func init() {
//	    registry.RegisterValidator("email", &EmailValidator{})
//	}
//
// Parameters:
//   - name: Unique identifier for this validator (must be non-empty)
//   - strategy: ValidatorStrategy implementation (must not be nil)
//
// Panics if name is empty or strategy is nil.
func RegisterValidator(name string, strategy api.ValidatorStrategy) {
	if strategy == nil {
		panic("registry: validator strategy cannot be nil")
	}

	RegisterProcessor(name, func() processor.ProcessorHandler {
		return processor.NewValidatorWrapper(strategy)
	})
}

// RegisterTransformer registers a TransformerStrategy by automatically wrapping it
// in a TransformWrapper. This is the recommended way to register transformers as it
// provides type safety and reduces boilerplate.
//
// The transformer will be automatically wrapped to implement ProcessorHandler
// and participate in the Chain of Responsibility pattern.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	type UpperCaseTransformer struct{}
//
//	func (t *UpperCaseTransformer) Transform(row map[string]interface{}) map[string]interface{} {
//	    result := make(map[string]interface{})
//	    for k, v := range row {
//	        if str, ok := v.(string); ok {
//	            result[k] = strings.ToUpper(str)
//	        } else {
//	            result[k] = v
//	        }
//	    }
//	    return result
//	}
//
//	func init() {
//	    registry.RegisterTransformer("uppercase", &UpperCaseTransformer{})
//	}
//
// Parameters:
//   - name: Unique identifier for this transformer (must be non-empty)
//   - strategy: TransformerStrategy implementation (must not be nil)
//
// Panics if name is empty or strategy is nil.
func RegisterTransformer(name string, strategy api.TransformerStrategy) {
	if strategy == nil {
		panic("registry: transformer strategy cannot be nil")
	}

	RegisterProcessor(name, func() processor.ProcessorHandler {
		return processor.NewTransformWrapper(strategy)
	})
}

// ErrProcessorNotFound is returned when a requested processor is not registered.
type ErrProcessorNotFound struct {
	Name string
}

func (e *ErrProcessorNotFound) Error() string {
	return fmt.Sprintf("processor not found: %s (available: %v)", e.Name, ListProcessors())
}

// ErrEmptyProcessorName is returned when an empty string is provided as processor name.
var ErrEmptyProcessorName = fmt.Errorf("processor name cannot be empty")
