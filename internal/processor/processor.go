package processor

// ProcessorHandler defines the interface for all processors in the chain.
// Each processor can process data and optionally pass it to the next processor.
type ProcessorHandler interface {
	SetNext(next ProcessorHandler)
	Process(data []map[string]interface{}) ([]map[string]interface{}, error)
}
