package engine

import (
	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

type ReportEngine struct {
	Provider  provider.ProviderStrategy
	Processor processor.ProcessorHandler
	Formatter formatter.FormatStrategy
	Output    output.OutputStrategy
}

func (r *ReportEngine) Run() error {
	// 1. Fetch data
	data, err := r.Provider.Fetch()
	if err != nil {
		return err
	}

	// 2. Process chain
	processed, err := r.Processor.Process(data)
	if err != nil {
		return err
	}

	// 3. Format
	formatted, err := r.Formatter.Format(processed)
	if err != nil {
		return err
	}

	// 4. Output
	return r.Output.Send(formatted)
}
