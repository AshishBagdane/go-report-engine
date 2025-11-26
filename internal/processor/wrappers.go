package processor

import (
	"context"

	"github.com/AshishBagdane/report-engine/pkg/api"
)

// --- Filter Wrapper ---

// FilterWrapper wraps a FilterStrategy to provide processor chain integration.
// It filters records based on the user-defined Keep() logic with context support.
//
// Context handling:
//   - Checks ctx.Done() before filtering
//   - Checks ctx.Done() periodically during iteration (every 100 rows)
//   - Returns ctx.Err() if canceled mid-processing
//   - Propagates context to next processor
//
// Thread-safe: Yes, if the underlying FilterStrategy is thread-safe.
type FilterWrapper struct {
	BaseProcessor // Handles SetNext and chain traversal
	strategy      api.FilterStrategy
}

// NewFilterWrapper creates a new FilterWrapper with the given strategy.
func NewFilterWrapper(s api.FilterStrategy) *FilterWrapper {
	return &FilterWrapper{strategy: s}
}

// Configure implements api.Configurable if the underlying strategy is configurable.
// This allows the strategy to receive configuration parameters.
func (f *FilterWrapper) Configure(params map[string]string) error {
	if configurable, ok := f.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

// Process filters data through the strategy and passes result to next processor.
// Records are filtered based on the FilterStrategy.Keep() method.
//
// Context handling:
//   - Checks context before starting
//   - Checks context every 100 rows for responsiveness
//   - Returns partial results with context error if canceled
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - data: Input data to filter
//
// Returns:
//   - []map[string]interface{}: Filtered data
//   - error: ctx.Err() if context canceled, or error from next processor
func (f *FilterWrapper) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check context before starting expensive operation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := make([]map[string]interface{}, 0, len(data))

	for i, row := range data {
		// Check for cancellation periodically (every 100 rows)
		// Balance between performance and cancellation responsiveness
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				// Return partial results with context error
				return result, ctx.Err()
			default:
			}
		}

		// Apply filter strategy
		if f.strategy.Keep(row) {
			result = append(result, row)
		}
	}

	// Pass filtered data to next processor with context
	return f.BaseProcessor.Process(ctx, result)
}

// --- Validator Wrapper ---

// ValidatorWrapper wraps a ValidatorStrategy to provide processor chain integration.
// It validates each record and fails fast on the first validation error.
//
// Context handling:
//   - Checks ctx.Done() before validation
//   - Checks ctx.Done() periodically during iteration (every 100 rows)
//   - Returns ctx.Err() if canceled mid-validation
//
// Thread-safe: Yes, if the underlying ValidatorStrategy is thread-safe.
type ValidatorWrapper struct {
	BaseProcessor
	strategy api.ValidatorStrategy
}

// NewValidatorWrapper creates a new ValidatorWrapper with the given strategy.
func NewValidatorWrapper(s api.ValidatorStrategy) *ValidatorWrapper {
	return &ValidatorWrapper{strategy: s}
}

// Configure implements api.Configurable if the underlying strategy is configurable.
func (v *ValidatorWrapper) Configure(params map[string]string) error {
	if configurable, ok := v.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

// Process validates each record and fails fast on first error.
// All records must pass validation for processing to continue.
//
// Context handling:
//   - Checks context before starting validation
//   - Checks context every 100 rows
//   - Returns context error with validated count
//   - Propagates context to next processor
//
// Fail-fast behavior: The first validation error stops processing
// and returns immediately.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - data: Input data to validate
//
// Returns:
//   - []map[string]interface{}: Validated data (same as input if all valid)
//   - error: Validation error, ctx.Err(), or error from next processor
func (v *ValidatorWrapper) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check context before starting validation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	for i, row := range data {
		// Check for cancellation periodically (every 100 rows)
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				// Return context error with count of validated rows
				return nil, ctx.Err()
			default:
			}
		}

		// Validate row - fail fast on first error
		if err := v.strategy.Validate(row); err != nil {
			return nil, err
		}
	}

	// All records validated successfully
	// Pass to next processor with context
	return v.BaseProcessor.Process(ctx, data)
}

// --- Transformer Wrapper ---

// TransformWrapper wraps a TransformerStrategy to provide processor chain integration.
// It transforms each record using the user-defined Transform() logic.
//
// Context handling:
//   - Checks ctx.Done() before transformation
//   - Checks ctx.Done() periodically during iteration (every 100 rows)
//   - Returns partial results with ctx.Err() if canceled
//   - Propagates context to next processor
//
// Thread-safe: Yes, if the underlying TransformerStrategy is thread-safe.
type TransformWrapper struct {
	BaseProcessor
	strategy api.TransformerStrategy
}

// NewTransformWrapper creates a new TransformWrapper with the given strategy.
func NewTransformWrapper(s api.TransformerStrategy) *TransformWrapper {
	return &TransformWrapper{strategy: s}
}

// Configure implements api.Configurable if the underlying strategy is configurable.
func (t *TransformWrapper) Configure(params map[string]string) error {
	if configurable, ok := t.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

// Process transforms each record using the strategy and passes result to next processor.
//
// Context handling:
//   - Checks context before starting transformation
//   - Checks context every 100 rows
//   - Returns partial results if canceled
//   - Propagates context to next processor
//
// Pre-allocates result slice for efficiency.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - data: Input data to transform
//
// Returns:
//   - []map[string]interface{}: Transformed data
//   - error: ctx.Err() if context canceled, or error from next processor
func (t *TransformWrapper) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check context before starting transformation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Pre-allocate result slice for efficiency
	result := make([]map[string]interface{}, len(data))

	for i, row := range data {
		// Check for cancellation periodically (every 100 rows)
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				// Return partial results with context error
				return result[:i], ctx.Err()
			default:
			}
		}

		// Apply transformation strategy
		result[i] = t.strategy.Transform(row)
	}

	// Pass transformed data to next processor with context
	return t.BaseProcessor.Process(ctx, result)
}
