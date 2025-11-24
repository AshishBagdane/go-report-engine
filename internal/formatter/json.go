package formatter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/AshishBagdane/report-engine/internal/logging"
)

// JSONFormatter formats data as indented JSON.
type JSONFormatter struct {
	logger *logging.Logger
}

func NewJSONFormatter() FormatStrategy {
	return &JSONFormatter{}
}

// WithLogger sets a custom logger for the formatter.
func (j *JSONFormatter) WithLogger(logger *logging.Logger) *JSONFormatter {
	j.logger = logger
	return j
}

// getLogger returns the formatter's logger, creating a default one if needed.
func (j *JSONFormatter) getLogger() *logging.Logger {
	if j.logger == nil {
		j.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "formatter.json",
		})
	}
	return j.logger
}

func (j *JSONFormatter) Format(data []map[string]interface{}) ([]byte, error) {
	logger := j.getLogger()
	ctx := context.Background()
	startTime := time.Now()
	recordCount := len(data)

	logger.InfoContext(ctx, "formatting starting",
		"formatter_type", "json",
		"record_count", recordCount,
	)

	result, err := json.MarshalIndent(data, "", "  ")
	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorContext(ctx, "formatting failed",
			"error", err,
			"formatter_type", "json",
			"record_count", recordCount,
			"duration_ms", duration.Milliseconds(),
		)
		return nil, err
	}

	outputSize := len(result)
	logger.InfoContext(ctx, "formatting completed",
		"formatter_type", "json",
		"record_count", recordCount,
		"output_size_bytes", outputSize,
		"duration_ms", duration.Milliseconds(),
	)

	if outputSize > 1024*1024 {
		logger.WarnContext(ctx, "large output generated",
			"output_size_bytes", outputSize,
			"output_size_mb", outputSize/1024/1024,
		)
	}

	return result, nil
}
