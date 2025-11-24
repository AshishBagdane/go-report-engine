// Package processor implements the Chain of Responsibility pattern for data processing.
package processor

import (
	"context"
	"time"

	"github.com/AshishBagdane/report-engine/internal/logging"
)

// BaseProcessor provides the chain traversal logic using the Template Method pattern.
// Concrete processors can embed this to get automatic chain handling.
type BaseProcessor struct {
	next   ProcessorHandler
	logger *logging.Logger
}

// SetNext sets the next processor in the chain.
func (b *BaseProcessor) SetNext(handler ProcessorHandler) {
	b.next = handler
}

// WithLogger sets a custom logger for the processor.
func (b *BaseProcessor) WithLogger(logger *logging.Logger) *BaseProcessor {
	b.logger = logger
	return b
}

// getLogger returns the processor's logger, creating a default one if needed.
func (b *BaseProcessor) getLogger() *logging.Logger {
	if b.logger == nil {
		b.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "processor.base",
		})
	}
	return b.logger
}

// Process passes data to the next processor if one exists, otherwise returns data as-is.
// This implements the Template Method pattern for chain traversal.
func (b *BaseProcessor) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	logger := b.getLogger()
	ctx := context.Background()

	if b.next != nil {
		startTime := time.Now()
		inputCount := len(data)

		logger.DebugContext(ctx, "passing to next processor",
			"input_records", inputCount,
		)

		result, err := b.next.Process(data)
		duration := time.Since(startTime)

		if err != nil {
			logger.ErrorContext(ctx, "next processor failed",
				"error", err,
				"input_records", inputCount,
				"duration_ms", duration.Milliseconds(),
			)
			return nil, err
		}

		outputCount := len(result)
		logger.DebugContext(ctx, "next processor completed",
			"input_records", inputCount,
			"output_records", outputCount,
			"duration_ms", duration.Milliseconds(),
		)

		return result, nil
	}

	logger.DebugContext(ctx, "end of chain reached",
		"record_count", len(data),
	)

	return data, nil
}
