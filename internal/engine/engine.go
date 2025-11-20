// Package engine provides the core report generation engine that orchestrates
// the data pipeline from provider through processors, formatter, and output.
package engine

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/errors"
	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

// ReportEngine orchestrates the report generation pipeline.
// It coordinates data flow through four main stages:
//  1. Provider - Fetches raw data
//  2. Processor - Processes and transforms data
//  3. Formatter - Formats data into desired output format
//  4. Output - Delivers formatted data to destination
//
// Each stage is pluggable via strategy interfaces, allowing for
// flexible configuration and testing.
type ReportEngine struct {
	// Provider fetches raw data from the data source
	Provider provider.ProviderStrategy

	// Processor processes data through a chain of transformations
	Processor processor.ProcessorHandler

	// Formatter converts processed data into the desired output format
	Formatter formatter.FormatStrategy

	// Output delivers the formatted data to its destination
	Output output.OutputStrategy
}

// Run executes the complete report generation pipeline.
// It orchestrates the four stages in sequence, wrapping any errors
// with context about which stage failed and what data was being processed.
//
// The pipeline stages are:
//  1. Fetch data from provider
//  2. Process data through processor chain
//  3. Format data into output format
//  4. Send formatted data to output
//
// If any stage fails, the pipeline stops immediately and returns
// a wrapped error with full context for debugging.
//
// Example:
//
//	engine := &ReportEngine{
//	    Provider:  provider.NewMockProvider(),
//	    Processor: &processor.BaseProcessor{},
//	    Formatter: formatter.NewJSONFormatter(),
//	    Output:    output.NewConsoleOutput(),
//	}
//
//	if err := engine.Run(); err != nil {
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
	// Validate that all required components are present
	if err := r.validate(); err != nil {
		return err
	}

	// Stage 1: Fetch data from provider
	data, err := r.fetchData()
	if err != nil {
		return err
	}

	// Stage 2: Process data through processor chain
	processed, err := r.processData(data)
	if err != nil {
		return err
	}

	// Stage 3: Format data
	formatted, err := r.formatData(processed)
	if err != nil {
		return err
	}

	// Stage 4: Output data
	if err := r.outputData(formatted); err != nil {
		return err
	}

	return nil
}

// validate ensures all required components are present.
// This prevents runtime panics and provides clear error messages.
func (r *ReportEngine) validate() error {
	if r.Provider == nil {
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("provider is required but not set")
	}
	if r.Processor == nil {
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("processor is required but not set")
	}
	if r.Formatter == nil {
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("formatter is required but not set")
	}
	if r.Output == nil {
		return errors.NewErrorContext(errors.ComponentEngine, "validate").
			WithType(errors.ErrorTypeConfiguration).
			New("output is required but not set")
	}
	return nil
}

// fetchData retrieves raw data from the provider with error wrapping.
func (r *ReportEngine) fetchData() ([]map[string]interface{}, error) {
	data, err := r.Provider.Fetch()
	if err != nil {
		// Wrap the error with engine context
		// The provider should have already wrapped it with provider context
		return nil, errors.Wrap(errors.ComponentEngine, "fetch_stage", err)
	}

	// Validate that we got some data
	if data == nil {
		return nil, errors.NewErrorContext(errors.ComponentEngine, "fetch_stage").
			WithType(errors.ErrorTypePermanent).
			New("provider returned nil data")
	}

	return data, nil
}

// processData runs data through the processor chain with error wrapping.
func (r *ReportEngine) processData(data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Track input record count for error context
	inputCount := len(data)

	processed, err := r.Processor.Process(data)
	if err != nil {
		// Wrap the error with engine context and add record count
		return nil, errors.NewErrorContext(errors.ComponentEngine, "process_stage").
			WithContext("input_records", inputCount).
			Wrap(err)
	}

	// Validate that we got processed data
	if processed == nil {
		return nil, errors.NewErrorContext(errors.ComponentEngine, "process_stage").
			WithType(errors.ErrorTypePermanent).
			WithContext("input_records", inputCount).
			New("processor returned nil data")
	}

	return processed, nil
}

// formatData converts processed data into the desired output format with error wrapping.
func (r *ReportEngine) formatData(data []map[string]interface{}) ([]byte, error) {
	// Track record count for error context
	recordCount := len(data)

	formatted, err := r.Formatter.Format(data)
	if err != nil {
		// Wrap the error with engine context and add record count
		return nil, errors.NewErrorContext(errors.ComponentEngine, "format_stage").
			WithContext("record_count", recordCount).
			Wrap(err)
	}

	// Validate that we got formatted data
	if formatted == nil {
		return nil, errors.NewErrorContext(errors.ComponentEngine, "format_stage").
			WithType(errors.ErrorTypePermanent).
			WithContext("record_count", recordCount).
			New("formatter returned nil data")
	}

	// Add size information to context for debugging
	if len(formatted) == 0 {
		return nil, errors.NewErrorContext(errors.ComponentEngine, "format_stage").
			WithType(errors.ErrorTypePermanent).
			WithContext("record_count", recordCount).
			New("formatter returned empty data")
	}

	return formatted, nil
}

// outputData sends formatted data to the output destination with error wrapping.
func (r *ReportEngine) outputData(data []byte) error {
	// Track data size for error context
	dataSize := len(data)

	err := r.Output.Send(data)
	if err != nil {
		// Wrap the error with engine context and add data size
		return errors.NewErrorContext(errors.ComponentEngine, "output_stage").
			WithContext("data_size", dataSize).
			Wrap(err)
	}

	return nil
}

// RunWithRecovery executes the pipeline with panic recovery.
// This is useful in production environments where you want to convert
// panics into errors rather than crashing the application.
//
// If a panic occurs, it's converted to an error with type ErrorTypePermanent
// and includes the panic value and stack trace in the context.
//
// Example:
//
//	err := engine.RunWithRecovery()
//	if err != nil {
//	    log.Printf("Pipeline failed: %v", err)
//	    // Send alert
//	}
//
// Returns:
//   - error: nil on success, or an error with panic information on panic
func (r *ReportEngine) RunWithRecovery() (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			// Convert panic to error
			err = errors.NewErrorContext(errors.ComponentEngine, "run").
				WithType(errors.ErrorTypePermanent).
				WithContext("panic", fmt.Sprintf("%v", rec)).
				New("pipeline panicked during execution")
		}
	}()

	return r.Run()
}
