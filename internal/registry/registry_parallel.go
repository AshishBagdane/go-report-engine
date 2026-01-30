package registry

import (
	"github.com/AshishBagdane/go-report-engine/internal/processor"
)

// RegisterParallelProcessor registers a processor wrapped in parallel execution mode.
// This is a convenience function that automatically wraps any registered processor
// with ParallelProcessor for concurrent execution.
//
// The parallel processor will:
//   - Split data into chunks for concurrent processing
//   - Execute the wrapped processor on chunks in parallel
//   - Reassemble results in original order
//   - Use default configuration (runtime.NumCPU() workers, auto chunk size)
//
// Configuration can be customized via the Configure() method after retrieval.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// Register a filter that will run in parallel
//	filter := &ExpensiveFilter{}
//	registry.RegisterParallelProcessor("parallel_filter", func() processor.ProcessorHandler {
//	    return processor.NewFilterWrapper(filter)
//	})
//
//	// Or wrap an existing registered processor
//	registry.RegisterProcessor("expensive_transform", func() processor.ProcessorHandler {
//	    return &ExpensiveTransform{}
//	})
//	registry.RegisterParallelProcessor("parallel_expensive_transform", func() processor.ProcessorHandler {
//	    base, _ := registry.GetProcessor("expensive_transform")
//	    return processor.NewParallelProcessor(base)
//	})
//
// Parameters:
//   - name: Unique identifier for this parallel processor (must be non-empty)
//   - factory: Function that creates the processor to be wrapped (must not be nil)
//
// Panics if name is empty or factory is nil.
func RegisterParallelProcessor(name string, factory ProcessorFactory) {
	if name == "" {
		panic("registry: processor name cannot be empty")
	}
	if factory == nil {
		panic("registry: processor factory cannot be nil")
	}

	// Wrap the factory to create a parallel processor
	parallelFactory := func() processor.ProcessorHandler {
		baseProcessor := factory()
		return processor.NewParallelProcessor(baseProcessor)
	}

	RegisterProcessor(name, parallelFactory)
}

// RegisterParallelProcessorWithConfig registers a parallel processor with custom configuration.
// This allows fine-grained control over worker count, chunk size, and other parallel settings.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	config := processor.ParallelConfig{
//	    Workers:      16,
//	    ChunkSize:    500,
//	    MinChunkSize: 50,
//	}
//
//	registry.RegisterParallelProcessorWithConfig("high_performance_filter", config, func() processor.ProcessorHandler {
//	    return processor.NewFilterWrapper(&CPUIntensiveFilter{})
//	})
//
// Parameters:
//   - name: Unique identifier for this parallel processor (must be non-empty)
//   - config: Parallel processing configuration
//   - factory: Function that creates the processor to be wrapped (must not be nil)
//
// Panics if name is empty or factory is nil.
func RegisterParallelProcessorWithConfig(name string, config processor.ParallelConfig, factory ProcessorFactory) {
	if name == "" {
		panic("registry: processor name cannot be empty")
	}
	if factory == nil {
		panic("registry: processor factory cannot be nil")
	}

	// Wrap the factory to create a configured parallel processor
	parallelFactory := func() processor.ProcessorHandler {
		baseProcessor := factory()
		return processor.NewParallelProcessorWithConfig(baseProcessor, config)
	}

	RegisterProcessor(name, parallelFactory)
}

// RegisterParallelFilter is a convenience function that registers a FilterStrategy
// wrapped in a FilterWrapper and then wrapped in ParallelProcessor for concurrent execution.
//
// This combines the benefits of type-safe filter registration with parallel processing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	type CPUIntensiveFilter struct {
//	    ComplexLogic bool
//	}
//
//	func (f *CPUIntensiveFilter) Keep(row map[string]interface{}) bool {
//	    // Expensive filtering logic
//	    return expensiveComputation(row)
//	}
//
//	func init() {
//	    registry.RegisterParallelFilter("cpu_intensive_filter", &CPUIntensiveFilter{ComplexLogic: true})
//	}
//
// Parameters:
//   - name: Unique identifier for this parallel filter (must be non-empty)
//   - strategy: FilterStrategy implementation (must not be nil)
//
// Panics if name is empty or strategy is nil.
func RegisterParallelFilter(name string, strategy interface {
	Keep(map[string]interface{}) bool
}) {
	if strategy == nil {
		panic("registry: filter strategy cannot be nil")
	}

	RegisterParallelProcessor(name, func() processor.ProcessorHandler {
		return processor.NewFilterWrapper(strategy)
	})
}

// RegisterParallelValidator is a convenience function that registers a ValidatorStrategy
// wrapped in a ValidatorWrapper and then wrapped in ParallelProcessor for concurrent execution.
//
// This combines the benefits of type-safe validator registration with parallel processing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	type ComplexValidator struct {
//	    Rules []ValidationRule
//	}
//
//	func (v *ComplexValidator) Validate(row map[string]interface{}) error {
//	    // Expensive validation logic
//	    return complexValidation(row)
//	}
//
//	func init() {
//	    registry.RegisterParallelValidator("complex_validator", &ComplexValidator{})
//	}
//
// Parameters:
//   - name: Unique identifier for this parallel validator (must be non-empty)
//   - strategy: ValidatorStrategy implementation (must not be nil)
//
// Panics if name is empty or strategy is nil.
func RegisterParallelValidator(name string, strategy interface {
	Validate(map[string]interface{}) error
}) {
	if strategy == nil {
		panic("registry: validator strategy cannot be nil")
	}

	RegisterParallelProcessor(name, func() processor.ProcessorHandler {
		return processor.NewValidatorWrapper(strategy)
	})
}

// RegisterParallelTransformer is a convenience function that registers a TransformerStrategy
// wrapped in a TransformWrapper and then wrapped in ParallelProcessor for concurrent execution.
//
// This combines the benefits of type-safe transformer registration with parallel processing.
//
// Thread-safe: Yes. Can be called concurrently from multiple goroutines.
//
// Example:
//
//	type ExpensiveTransformer struct {
//	    ComplexLogic bool
//	}
//
//	func (t *ExpensiveTransformer) Transform(row map[string]interface{}) map[string]interface{} {
//	    // Expensive transformation logic
//	    return expensiveTransform(row)
//	}
//
//	func init() {
//	    registry.RegisterParallelTransformer("expensive_transformer", &ExpensiveTransformer{})
//	}
//
// Parameters:
//   - name: Unique identifier for this parallel transformer (must be non-empty)
//   - strategy: TransformerStrategy implementation (must not be nil)
//
// Panics if name is empty or strategy is nil.
func RegisterParallelTransformer(name string, strategy interface {
	Transform(map[string]interface{}) map[string]interface{}
}) {
	if strategy == nil {
		panic("registry: transformer strategy cannot be nil")
	}

	RegisterParallelProcessor(name, func() processor.ProcessorHandler {
		return processor.NewTransformWrapper(strategy)
	})
}
