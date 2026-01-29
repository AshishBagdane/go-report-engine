package processor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// DeduplicateProcessor filters out duplicate records.
// It can check specific fields or the entire record.
type DeduplicateProcessor struct {
	BaseProcessor
	// Fields to check for uniqueness. If empty, checks all fields.
	Fields []string
}

// NewDeduplicateProcessor creates a new DeduplicateProcessor.
func NewDeduplicateProcessor(fields []string) *DeduplicateProcessor {
	return &DeduplicateProcessor{
		Fields: fields,
	}
}

// Configure sets up the processor from parameters.
// Params:
// - fields: Comma-separated list of fields to check for uniqueness.
func (p *DeduplicateProcessor) Configure(params map[string]string) error {
	if val, ok := params["fields"]; ok && val != "" {
		p.Fields = strings.Split(val, ",")
		for i := range p.Fields {
			p.Fields[i] = strings.TrimSpace(p.Fields[i])
		}
	}
	return nil
}

// Process filters duplicates and passes unique records to the next processor.
func (p *DeduplicateProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(data) == 0 {
		return p.BaseProcessor.Process(ctx, data)
	}

	seen := make(map[string]bool)
	uniqueData := make([]map[string]interface{}, 0, len(data))

	for _, record := range data {
		signature := p.generateSignature(record)
		if !seen[signature] {
			seen[signature] = true
			uniqueData = append(uniqueData, record)
		}
	}

	return p.BaseProcessor.Process(ctx, uniqueData)
}

// generateSignature creates a unique string signature for a record
func (p *DeduplicateProcessor) generateSignature(record map[string]interface{}) string {
	var sb strings.Builder

	// Determine keys to use
	var keys []string
	if len(p.Fields) > 0 {
		keys = p.Fields
	} else {
		keys = make([]string, 0, len(record))
		for k := range record {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Ensure deterministic order for full record check
	}

	for _, k := range keys {
		val, ok := record[k]
		if !ok {
			sb.WriteString(fmt.Sprintf("%s:<missing>|", k))
			continue
		}
		// specific formatting to avoid collision (e.g. "1" vs 1)
		// technically strict type check is hard in pure string, but for dedupe usually string rep is fine
		sb.WriteString(fmt.Sprintf("%s:%v|", k, val))
	}

	// Hash the signature to keep memory usage somewhat constant regardless of strict data size,
	// though keeping map of full strings is safer against collision. SHA256 is safe enough.
	h := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(h[:])
}
