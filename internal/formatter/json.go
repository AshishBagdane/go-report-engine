package formatter

import (
	"encoding/json"
)

func NewJSONFormatter() FormatStrategy {
	return &JSONFormatter{}
}

type JSONFormatter struct{}

func (j *JSONFormatter) Format(data []map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
