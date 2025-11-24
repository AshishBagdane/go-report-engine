package output

import (
	"context"
	"fmt"
	"time"

	"github.com/AshishBagdane/report-engine/internal/logging"
)

// ConsoleOutput writes formatted data to stdout.
type ConsoleOutput struct {
	logger *logging.Logger
}

func NewConsoleOutput() OutputStrategy {
	return &ConsoleOutput{}
}

// WithLogger sets a custom logger for the output.
func (c *ConsoleOutput) WithLogger(logger *logging.Logger) *ConsoleOutput {
	c.logger = logger
	return c
}

// getLogger returns the output's logger, creating a default one if needed.
func (c *ConsoleOutput) getLogger() *logging.Logger {
	if c.logger == nil {
		c.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "output.console",
		})
	}
	return c.logger
}

func (c *ConsoleOutput) Send(data []byte) error {
	logger := c.getLogger()
	ctx := context.Background()
	startTime := time.Now()
	dataSize := len(data)

	logger.InfoContext(ctx, "send starting",
		"output_type", "console",
		"destination", "stdout",
		"data_size_bytes", dataSize,
	)

	fmt.Println(string(data))

	duration := time.Since(startTime)
	logger.InfoContext(ctx, "send completed",
		"output_type", "console",
		"destination", "stdout",
		"data_size_bytes", dataSize,
		"duration_ms", duration.Milliseconds(),
	)

	return nil
}
