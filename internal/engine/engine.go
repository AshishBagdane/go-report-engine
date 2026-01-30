// Package engine provides the core report generation engine that orchestrates
// the data pipeline from provider through processors, formatter, and output.
//
// This file adds comprehensive structured logging throughout the engine lifecycle.
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AshishBagdane/go-report-engine/internal/errors"
	"github.com/AshishBagdane/go-report-engine/internal/formatter"
	"github.com/AshishBagdane/go-report-engine/internal/health"
	"github.com/AshishBagdane/go-report-engine/internal/logging"
	"github.com/AshishBagdane/go-report-engine/internal/memory"
	"github.com/AshishBagdane/go-report-engine/internal/output"
	"github.com/AshishBagdane/go-report-engine/internal/processor"
	"github.com/AshishBagdane/go-report-engine/internal/provider"
)

// ReportEngine orchestrates the report generation pipeline with structured logging.
// It coordinates data flow through four main stages:
//  1. Provider - Fetches raw data
//  2. Processor - Processes and transforms data
//  3. Formatter - Formats data into desired output format
//  4. Output - Delivers formatted data to destination
//
// Each stage is pluggable via strategy interfaces and fully logged.
type ReportEngine struct {
	// Provider fetches raw data from the data source
	Provider provider.ProviderStrategy

	// Processor processes data through a chain of transformations
	Processor processor.ProcessorHandler

	// Formatter converts processed data into the desired output format
	Formatter formatter.FormatStrategy

	// Output delivers the formatted data to its destination
	Output output.OutputStrategy

	// logger provides structured logging for the engine
	logger *logging.Logger

	// ChunkSize is the number of records to process in each batch during streaming.
	// Default: 1000
	ChunkSize int

	// closeOnce ensures cleanup is performed exactly once
	closeOnce sync.Once
}

// WithChunkSize sets the chunk size for streaming operations.
func (r *ReportEngine) WithChunkSize(size int) *ReportEngine {
	r.ChunkSize = size
	return r
}

func (r *ReportEngine) getChunkSize() int {
	if r.ChunkSize <= 0 {
		return 1000
	}
	return r.ChunkSize
}

// WithLogger sets a custom logger for the engine.
// If not set, a default logger will be created on first use.
//
// Example:
//
//	logger := logging.NewLogger(logging.Config{
//	    Level: logging.LevelDebug,
//	    Format: logging.FormatJSON,
//	    Component: "engine",
//	})
//	engine.WithLogger(logger)
func (r *ReportEngine) WithLogger(logger *logging.Logger) *ReportEngine {
	r.logger = logger
	return r
}

// getLogger returns the engine's logger, creating a default one if needed.
func (r *ReportEngine) getLogger() *logging.Logger {
	if r.logger == nil {
		r.logger = logging.NewLogger(logging.Config{
			Level:     logging.LevelInfo,
			Format:    logging.FormatJSON,
			Component: "engine",
		})
	}
	return r.logger
}

// Run executes the complete report generation pipeline with comprehensive logging.
// It orchestrates the four stages in sequence, logging performance metrics and
// errors at each stage.
//
// The pipeline stages are:
//  1. Fetch data from provider
//  2. Process data through processor chain
//  3. Format data into output format
//  4. Send formatted data to output
//
// All stages are logged with:
//   - Start/completion messages
//   - Performance metrics (duration, record counts, data sizes)
//   - Error context (stage, counts, causes)
//
// Example:
//
//	ctx := logging.WithRequestID(context.Background(), "req-123")
//	if err := engine.RunWithContext(ctx); err != nil {
//	    if errors.IsRetryable(err) {
//	        // Retry logic
//	    } else {
//	        log.Fatal(err)
//	    }
//	}
//
// Returns:
//   - error: nil on success, or an errors.EngineError with full context on failure
func (r *ReportEngine) Run() error {
	return r.RunWithContext(context.Background())
}

// RunWithContext executes the pipeline with context for cancellation and tracing.
// It supports request IDs, correlation IDs, and cancellation via context.
//
// Example:
//
//	ctx := logging.WithRequestID(context.Background(), "req-abc-123")
//	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
//	defer cancel()
//
//	if err := engine.RunWithContext(ctx); err != nil {
//	    logger.ErrorContext(ctx, "pipeline failed", "error", err)
//	}
func (r *ReportEngine) RunWithContext(ctx context.Context) error {
	logger := r.getLogger()
	startTime := time.Now()

	logger.InfoContext(ctx, "engine starting",
		"stage", "validation",
	)

	// Validate that all required components are present
	if err := r.validate(); err != nil {
		logger.ErrorContext(ctx, "engine validation failed",
			"error", err,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return err
	}

	logger.DebugContext(ctx, "engine validation passed")

	// Check for streaming support
	streamingProvider, okProvider := r.Provider.(provider.StreamingProviderStrategy)
	streamingFormatter, okFormatter := r.Formatter.(formatter.StreamingFormatterStrategy)
	streamingOutput, okOutput := r.Output.(output.StreamingOutputStrategy)

	if okProvider && okFormatter && okOutput {
		logger.InfoContext(ctx, "executing streaming pipeline")
		return r.runStreamingPipeline(ctx, streamingProvider, streamingFormatter, streamingOutput)
	}

	logger.InfoContext(ctx, "executing batch pipeline")

	// Stage 1: Fetch data from provider
	data, err := r.fetchDataWithContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "pipeline failed at fetch stage",
			"error", err,
			"stage", "fetch",
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return err
	}

	// Stage 2: Process data through processor chain
	processed, err := r.processDataWithContext(ctx, data)
	if err != nil {
		logger.ErrorContext(ctx, "pipeline failed at process stage",
			"error", err,
			"stage", "process",
			"input_records", len(data),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return err
	}

	// Stage 3: Format data
	formatted, err := r.formatDataWithContext(ctx, processed)
	if err != nil {
		logger.ErrorContext(ctx, "pipeline failed at format stage",
			"error", err,
			"stage", "format",
			"record_count", len(processed),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return err
	}

	// Stage 4: Output data
	if err := r.outputDataWithContext(ctx, formatted); err != nil {
		logger.ErrorContext(ctx, "pipeline failed at output stage",
			"error", err,
			"stage", "output",
			"data_size_bytes", len(formatted),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return err
	}

	// Success - log completion metrics
	duration := time.Since(startTime)
	logger.InfoContext(ctx, "engine completed successfully",
		"total_duration_ms", duration.Milliseconds(),
		"input_records", len(data),
		"output_records", len(processed),
		"output_size_bytes", len(formatted),
	)

	return nil
}

// validate ensures all required components are present.
// This prevents runtime panics and provides clear error messages.
func (r *ReportEngine) validate() error {
	logger := r.getLogger()

	if r.Provider == nil {
		logger.Error("validation failed: provider is nil")
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("provider is required but not set")
	}
	if r.Processor == nil {
		logger.Error("validation failed: processor is nil")
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("processor is required but not set")
	}
	if r.Formatter == nil {
		logger.Error("validation failed: formatter is nil")
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("formatter is required but not set")
	}
	if r.Output == nil {
		logger.Error("validation failed: output is nil")
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("output is required but not set")
	}

	logger.Debug("validation passed: all components present")
	return nil
}

// fetchDataWithContext retrieves raw data from the provider with logging.
func (r *ReportEngine) fetchDataWithContext(ctx context.Context) ([]map[string]interface{}, error) {
	logger := r.getLogger()
	startTime := time.Now()

	logger.InfoContext(ctx, "fetch stage starting")

	// UPDATED: Pass context to provider
	data, err := r.Provider.Fetch(ctx)
	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorContext(ctx, "fetch stage failed",
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		// Wrap the error with engine context
		return nil, errors.Wrap(errors.ComponentEngine, "fetch_stage", err)
	}

	// Validate that we got some data
	if data == nil {
		logger.WarnContext(ctx, "provider returned nil data")
		return nil, errors.NewErrorContext(errors.ComponentEngine, "fetch_stage").
			WithType(errors.ErrorTypePermanent).
			New("provider returned nil data")
	}

	recordCount := len(data)
	logger.InfoContext(ctx, "fetch stage completed",
		"duration_ms", duration.Milliseconds(),
		"record_count", recordCount,
	)

	// Log warning if no records fetched
	if recordCount == 0 {
		logger.WarnContext(ctx, "provider returned zero records")
	}

	return data, nil
}

// processDataWithContext runs data through the processor chain with logging.
func (r *ReportEngine) processDataWithContext(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	logger := r.getLogger()
	startTime := time.Now()
	inputCount := len(data)

	logger.InfoContext(ctx, "process stage starting",
		"input_records", inputCount,
	)

	// UPDATED: Pass context to processor
	processed, err := r.Processor.Process(ctx, data)
	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorContext(ctx, "process stage failed",
			"error", err,
			"input_records", inputCount,
			"duration_ms", duration.Milliseconds(),
		)
		// Wrap the error with engine context and add record count
		return nil, errors.NewErrorContext(errors.ComponentEngine, "process_stage").
			WithContext("input_records", inputCount).
			Wrap(err)
	}

	// Validate that we got processed data
	if processed == nil {
		logger.WarnContext(ctx, "processor returned nil data",
			"input_records", inputCount,
		)
		return nil, errors.NewErrorContext(errors.ComponentEngine, "process_stage").
			WithType(errors.ErrorTypePermanent).
			WithContext("input_records", inputCount).
			New("processor returned nil data")
	}

	outputCount := len(processed)
	filteredCount := inputCount - outputCount

	logger.InfoContext(ctx, "process stage completed",
		"duration_ms", duration.Milliseconds(),
		"input_records", inputCount,
		"output_records", outputCount,
		"filtered_records", filteredCount,
	)

	// Log warning if all records were filtered
	if outputCount == 0 && inputCount > 0 {
		logger.WarnContext(ctx, "all records filtered by processor",
			"input_records", inputCount,
		)
	}

	return processed, nil
}

// formatDataWithContext converts processed data into the desired output format with logging.
func (r *ReportEngine) formatDataWithContext(ctx context.Context, data []map[string]interface{}) ([]byte, error) {
	logger := r.getLogger()
	startTime := time.Now()
	recordCount := len(data)

	logger.InfoContext(ctx, "format stage starting",
		"record_count", recordCount,
	)

	// UPDATED: Pass context to formatter
	formatted, err := r.Formatter.Format(ctx, data)
	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorContext(ctx, "format stage failed",
			"error", err,
			"record_count", recordCount,
			"duration_ms", duration.Milliseconds(),
		)
		// Wrap the error with engine context and add record count
		return nil, errors.NewErrorContext(errors.ComponentEngine, "format_stage").
			WithContext("record_count", recordCount).
			Wrap(err)
	}

	// Validate that we got formatted data
	if formatted == nil {
		logger.WarnContext(ctx, "formatter returned nil data",
			"record_count", recordCount,
		)
		return nil, errors.NewErrorContext(errors.ComponentEngine, "format_stage").
			WithType(errors.ErrorTypePermanent).
			WithContext("record_count", recordCount).
			New("formatter returned nil data")
	}

	// Add size information to context for debugging
	dataSize := len(formatted)
	if dataSize == 0 {
		logger.WarnContext(ctx, "formatter returned empty data",
			"record_count", recordCount,
		)
		return nil, errors.NewErrorContext(errors.ComponentEngine, "format_stage").
			WithType(errors.ErrorTypePermanent).
			WithContext("record_count", recordCount).
			New("formatter returned empty data")
	}

	logger.InfoContext(ctx, "format stage completed",
		"duration_ms", duration.Milliseconds(),
		"record_count", recordCount,
		"output_size_bytes", dataSize,
		"bytes_per_record", dataSize/max(recordCount, 1),
	)

	return formatted, nil
}

// outputDataWithContext sends formatted data to the output destination with logging.
func (r *ReportEngine) outputDataWithContext(ctx context.Context, data []byte) error {
	logger := r.getLogger()
	startTime := time.Now()
	dataSize := len(data)

	logger.InfoContext(ctx, "output stage starting",
		"data_size_bytes", dataSize,
	)

	// UPDATED: Pass context to output
	err := r.Output.Send(ctx, data)
	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorContext(ctx, "output stage failed",
			"error", err,
			"data_size_bytes", dataSize,
			"duration_ms", duration.Milliseconds(),
		)
		// Wrap the error with engine context and add data size
		return errors.NewErrorContext(errors.ComponentEngine, "output_stage").
			WithContext("data_size", dataSize).
			Wrap(err)
	}

	logger.InfoContext(ctx, "output stage completed",
		"duration_ms", duration.Milliseconds(),
		"data_size_bytes", dataSize,
		"throughput_bytes_per_ms", dataSize/max(int(duration.Milliseconds()), 1),
	)

	return nil
}

// RunWithRecovery executes the pipeline with panic recovery and logging.
// This is useful in production environments where you want to convert
// panics into errors rather than crashing the application.
//
// If a panic occurs, it's converted to an error with type ErrorTypePermanent
// and includes the panic value in the context and logs.
//
// Example:
//
//	ctx := logging.WithRequestID(context.Background(), "req-123")
//	err := engine.RunWithRecoveryContext(ctx)
//	if err != nil {
//	    log.Printf("Pipeline failed: %v", err)
//	    // Send alert
//	}
//
// Returns:
//   - error: nil on success, or an error with panic information on panic
func (r *ReportEngine) RunWithRecovery() (err error) {
	return r.RunWithRecoveryContext(context.Background())
}

// RunWithRecoveryContext executes the pipeline with panic recovery and context.
func (r *ReportEngine) RunWithRecoveryContext(ctx context.Context) (err error) {
	logger := r.getLogger()

	defer func() {
		if rec := recover(); rec != nil {
			logger.ErrorContext(ctx, "engine panicked during execution",
				"panic", fmt.Sprintf("%v", rec),
			)

			// Convert panic to error
			err = errors.NewErrorContext(errors.ComponentEngine, "run").
				WithType(errors.ErrorTypePermanent).
				WithContext("panic", fmt.Sprintf("%v", rec)).
				New("pipeline panicked during execution")
		}
	}()

	return r.RunWithContext(ctx)
}

// Health checks the health of the engine and its components.
func (r *ReportEngine) Health(ctx context.Context) map[string]health.Result {
	results := make(map[string]health.Result)

	// Check Provider
	if checker, ok := r.Provider.(health.Checker); ok {
		res, err := checker.CheckHealth(ctx)
		if err != nil && res.Status == "" {
			res.Status = health.StatusDown
			res.Error = err.Error()
		}
		results["provider"] = res
	} else {
		results["provider"] = health.Result{Status: health.StatusUp, Details: map[string]interface{}{"note": "health check not implemented"}}
	}

	// Check Output
	if checker, ok := r.Output.(health.Checker); ok {
		res, err := checker.CheckHealth(ctx)
		if err != nil && res.Status == "" {
			res.Status = health.StatusDown
			res.Error = err.Error()
		}
		results["output"] = res
	} else {
		results["output"] = health.Result{Status: health.StatusUp, Details: map[string]interface{}{"note": "health check not implemented"}}
	}

	// Check Processor (usually stateless but technically could report health)
	if checker, ok := r.Processor.(health.Checker); ok {
		res, err := checker.CheckHealth(ctx)
		if err != nil && res.Status == "" {
			res.Status = health.StatusDown
			res.Error = err.Error()
		}
		results["processor"] = res
	}

	return results
}

// max returns the maximum of two integers (helper for Go versions < 1.21)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// runStreamingPipeline executes the pipeline in streaming mode.
func (r *ReportEngine) runStreamingPipeline(
	ctx context.Context,
	prov provider.StreamingProviderStrategy,
	fmttr formatter.StreamingFormatterStrategy,
	out output.StreamingOutputStrategy,
) error {
	logger := r.getLogger()
	startTime := time.Now()

	// Initialize output
	if err := out.Initialize(ctx); err != nil {
		logger.ErrorContext(ctx, "streaming: output initialization failed", "error", err)
		return errors.NewErrorContext(errors.ComponentOutput, "initialize").Wrap(err)
	}
	defer func() {
		if err := out.Close(ctx); err != nil {
			logger.WarnContext(ctx, "streaming: output close failed", "error", err)
		}
	}()

	// Start stream
	iterator, err := prov.Stream(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "streaming: provider stream failed", "error", err)
		return errors.NewErrorContext(errors.ComponentProvider, "stream").Wrap(err)
	}
	defer func() {
		if err := iterator.Close(); err != nil {
			logger.WarnContext(ctx, "streaming: iterator close failed", "error", err)
		}
	}()

	// Format Start
	startBytes, err := fmttr.FormatStart(ctx)
	if err != nil {
		return errors.NewErrorContext(errors.ComponentFormatter, "format_start").Wrap(err)
	}
	if err := out.WriteChunk(ctx, startBytes); err != nil {
		return errors.NewErrorContext(errors.ComponentOutput, "write_chunk").Wrap(err)
	}

	chunkSize := r.getChunkSize()
	buffer := make([]map[string]interface{}, 0, chunkSize)
	totalRecords := 0
	isFirstChunk := true

	// Iterate
	for iterator.Next() {
		buffer = append(buffer, iterator.Value())

		if len(buffer) >= chunkSize {
			if err := r.processAndWriteChunk(ctx, buffer, fmttr, out, &isFirstChunk); err != nil {
				return err
			}
			totalRecords += len(buffer)
			buffer = buffer[:0]
		}
	}
	if err := iterator.Err(); err != nil {
		logger.ErrorContext(ctx, "streaming: iterator error", "error", err)
		return errors.NewErrorContext(errors.ComponentProvider, "iterate").Wrap(err)
	}

	// Process remaining
	if len(buffer) > 0 {
		if err := r.processAndWriteChunk(ctx, buffer, fmttr, out, &isFirstChunk); err != nil {
			return err
		}
		totalRecords += len(buffer)
	}

	// Format End
	endBytes, err := fmttr.FormatEnd(ctx)
	if err != nil {
		return errors.NewErrorContext(errors.ComponentFormatter, "format_end").Wrap(err)
	}
	if err := out.WriteChunk(ctx, endBytes); err != nil {
		return errors.NewErrorContext(errors.ComponentOutput, "write_chunk").Wrap(err)
	}

	logger.InfoContext(ctx, "streaming pipeline completed",
		"total_records", totalRecords,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)
	return nil
}

func (r *ReportEngine) processAndWriteChunk(
	ctx context.Context,
	chunk []map[string]interface{},
	fmttr formatter.StreamingFormatterStrategy,
	out output.StreamingOutputStrategy,
	isFirstChunk *bool,
) error {
	// Release maps back to pool after processing
	defer func() {
		for _, m := range chunk {
			memory.PutMap(m)
		}
	}()

	// Process
	processed, err := r.Processor.Process(ctx, chunk)
	if err != nil {
		return errors.NewErrorContext(errors.ComponentProcessor, "process_chunk").Wrap(err)
	}
	if len(processed) == 0 {
		return nil
	}

	// Delimiter
	if !*isFirstChunk {
		if err := out.WriteChunk(ctx, []byte(",")); err != nil {
			return errors.NewErrorContext(errors.ComponentOutput, "write_delimiter").Wrap(err)
		}
	}
	*isFirstChunk = false

	// Format
	bytes, err := fmttr.FormatChunk(ctx, processed)
	if err != nil {
		return errors.NewErrorContext(errors.ComponentFormatter, "format_chunk").Wrap(err)
	}

	// Write
	if err := out.WriteChunk(ctx, bytes); err != nil {
		return errors.NewErrorContext(errors.ComponentOutput, "write_chunk").Wrap(err)
	}
	return nil
}
