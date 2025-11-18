package processor

type BaseProcessor struct {
	next ProcessorHandler
}

func (b *BaseProcessor) SetNext(next ProcessorHandler) {
	b.next = next
}

func (b *BaseProcessor) Process(data []map[string]interface{}) ([]map[string]interface{}, error) {
	if b.next != nil {
		return b.next.Process(data)
	}
	return data, nil
}
