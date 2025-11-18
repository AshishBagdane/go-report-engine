package main

import (
	"fmt"

	"github.com/AshishBagdane/report-engine/internal/engine"
	"github.com/AshishBagdane/report-engine/internal/formatter"
	"github.com/AshishBagdane/report-engine/internal/output"
	"github.com/AshishBagdane/report-engine/internal/processor"
	"github.com/AshishBagdane/report-engine/internal/provider"
)

func main() {
	mock := provider.NewMockProvider()
	processorChain := &processor.BaseProcessor{}
	jsonFmt := formatter.NewJSONFormatter()
	consoleOut := output.NewConsoleOutput()

	re := engine.ReportEngine{
		Provider:  mock,
		Processor: processorChain,
		Formatter: jsonFmt,
		Output:    consoleOut,
	}

	if err := re.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
