package processor

type ProcessorHandler interface {
	SetNext(next ProcessorHandler)
	Process(data []map[string]interface{}) ([]map[string]interface{}, error)
}
