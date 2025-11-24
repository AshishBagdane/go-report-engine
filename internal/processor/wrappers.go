package processor

import (
	"context"
	"time"

	"github.com/AshishBagdane/report-engine/internal/logging"
	"github.com/AshishBagdane/report-engine/pkg/api"
)

// FilterWrapper wraps a FilterStrategy with logging.
type FilterWrapper struct {
	BaseProcessor
	strategy api.FilterStrategy
	logger   *logging.Logger
}

func NewFilterWrapper(s api.FilterStrategy) *FilterWrapper {
	return &FilterWrapper{strategy: s}
}

func (f *FilterWrapper) WithLogger(logger *logging.Logger) *FilterWrapper {
	f.logger = logger
	return f
}

func (f *FilterWrapper) getLogger() *logging.Logger {
	if f.logger == nil {
		f.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "processor.filter",
		})
	}
	return f.logger
}

func (f *FilterWrapper) Configure(params map[string]string) error {
	if configurable, ok := f.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

func (f *FilterWrapper) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	logger := f.getLogger()
	ctx := context.Background()
	startTime := time.Now()
	inputCount := len(data)

	logger.InfoContext(ctx, "filter processing starting",
		"input_records", inputCount,
	)

	var result []map[string]interface{}
	for _, row := range data {
		if f.strategy.Keep(row) {
			result = append(result, row)
		}
	}

	duration := time.Since(startTime)
	outputCount := len(result)
	filteredCount := inputCount - outputCount

	logger.InfoContext(ctx, "filter processing completed",
		"input_records", inputCount,
		"output_records", outputCount,
		"filtered_records", filteredCount,
		"duration_ms", duration.Milliseconds(),
	)

	if outputCount == 0 && inputCount > 0 {
		logger.WarnContext(ctx, "filter removed all records",
			"input_records", inputCount,
		)
	}

	return f.BaseProcessor.Process(result)
}

// ValidatorWrapper wraps a ValidatorStrategy with logging.
type ValidatorWrapper struct {
	BaseProcessor
	strategy api.ValidatorStrategy
	logger   *logging.Logger
}

func NewValidatorWrapper(s api.ValidatorStrategy) *ValidatorWrapper {
	return &ValidatorWrapper{strategy: s}
}

func (v *ValidatorWrapper) WithLogger(logger *logging.Logger) *ValidatorWrapper {
	v.logger = logger
	return v
}

func (v *ValidatorWrapper) getLogger() *logging.Logger {
	if v.logger == nil {
		v.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "processor.validator",
		})
	}
	return v.logger
}

func (v *ValidatorWrapper) Configure(params map[string]string) error {
	if configurable, ok := v.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

func (v *ValidatorWrapper) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	logger := v.getLogger()
	ctx := context.Background()
	startTime := time.Now()
	recordCount := len(data)

	logger.InfoContext(ctx, "validation starting",
		"record_count", recordCount,
	)

	for i, row := range data {
		if err := v.strategy.Validate(row); err != nil {
			duration := time.Since(startTime)
			logger.ErrorContext(ctx, "validation failed",
				"error", err,
				"record_index", i,
				"total_records", recordCount,
				"duration_ms", duration.Milliseconds(),
			)
			return nil, err
		}
	}

	duration := time.Since(startTime)
	logger.InfoContext(ctx, "validation completed",
		"record_count", recordCount,
		"duration_ms", duration.Milliseconds(),
	)

	return v.BaseProcessor.Process(data)
}

// TransformWrapper wraps a TransformerStrategy with logging.
type TransformWrapper struct {
	BaseProcessor
	strategy api.TransformerStrategy
	logger   *logging.Logger
}

func NewTransformWrapper(s api.TransformerStrategy) *TransformWrapper {
	return &TransformWrapper{strategy: s}
}

func (t *TransformWrapper) WithLogger(logger *logging.Logger) *TransformWrapper {
	t.logger = logger
	return t
}

func (t *TransformWrapper) getLogger() *logging.Logger {
	if t.logger == nil {
		t.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "processor.transform",
		})
	}
	return t.logger
}

func (t *TransformWrapper) Configure(params map[string]string) error {
	if configurable, ok := t.strategy.(api.Configurable); ok {
		return configurable.Configure(params)
	}
	return nil
}

func (t *TransformWrapper) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	logger := t.getLogger()
	ctx := context.Background()
	startTime := time.Now()
	recordCount := len(data)

	logger.InfoContext(ctx, "transformation starting",
		"record_count", recordCount,
	)

	result := make([]map[string]interface{}, len(data))
	for i, row := range data {
		result[i] = t.strategy.Transform(row)
	}

	duration := time.Since(startTime)
	logger.InfoContext(ctx, "transformation completed",
		"record_count", recordCount,
		"duration_ms", duration.Milliseconds(),
	)

	return t.BaseProcessor.Process(result)
}
