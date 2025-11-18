package processor

import (
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// --- Filter Wrapper ---

type FilterWrapper struct {
	BaseProcessor // Handles SetNext and chain traversal
	strategy      api.FilterStrategy
}

func NewFilterWrapper(s api.FilterStrategy) *FilterWrapper {
	return &FilterWrapper{strategy: s}
}

// Configure checks if the underlying strategy is configurable and calls Configure on it.
func (f *FilterWrapper) Configure(params map[string]string) error {
	if configurable, ok := f.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

func (f *FilterWrapper) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, row := range data {
		if f.strategy.Keep(row) {
			result = append(result, row)
		}
	}
	return f.BaseProcessor.Process(result)
}

// --- Validator Wrapper ---

type ValidatorWrapper struct {
	BaseProcessor
	strategy api.ValidatorStrategy
}

func NewValidatorWrapper(s api.ValidatorStrategy) *ValidatorWrapper {
	return &ValidatorWrapper{strategy: s}
}

func (v *ValidatorWrapper) Configure(params map[string]string) error {
	if configurable, ok := v.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

func (v *ValidatorWrapper) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	for _, row := range data {
		if err := v.strategy.Validate(row); err != nil {
			return nil, err // Fail fast on validation error
		}
	}
	return v.BaseProcessor.Process(data)
}

// --- Transformer Wrapper ---

type TransformWrapper struct {
	BaseProcessor
	strategy api.TransformerStrategy
}

func NewTransformWrapper(s api.TransformerStrategy) *TransformWrapper {
	return &TransformWrapper{strategy: s}
}

func (t *TransformWrapper) Configure(params map[string]string) error {
	if configurable, ok := t.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

func (t *TransformWrapper) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(data))
	for i, row := range data {
		result[i] = t.strategy.Transform(row)
	}
	return t.BaseProcessor.Process(result)
}
