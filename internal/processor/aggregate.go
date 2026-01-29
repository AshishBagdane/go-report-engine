package processor

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// AggregateProcessor groups data and calculates aggregates.
type AggregateProcessor struct {
	BaseProcessor
	// GroupBy fields to group records by.
	GroupBy []string
	// Aggregates defines calculations: map[OutputField]Operation:InputField
	// Example: "total_sales": "sum:amount"
	// Operations: sum, count, min, max, avg
	Aggregates map[string]string
}

// NewAggregateProcessor creates a new AggregateProcessor.
func NewAggregateProcessor(groupBy []string, aggregates map[string]string) *AggregateProcessor {
	return &AggregateProcessor{
		GroupBy:    groupBy,
		Aggregates: aggregates,
	}
}

// Configure sets up the processor from parameters.
// Params:
// - group_by: Comma-separated list of fields to group by.
// - agg_<Output>: Operation spec, e.g. "agg_total" -> "sum:amount"
func (p *AggregateProcessor) Configure(params map[string]string) error {
	if val, ok := params["group_by"]; ok && val != "" {
		p.GroupBy = strings.Split(val, ",")
		for i := range p.GroupBy {
			p.GroupBy[i] = strings.TrimSpace(p.GroupBy[i])
		}
	}

	if p.Aggregates == nil {
		p.Aggregates = make(map[string]string)
	}

	for k, v := range params {
		if strings.HasPrefix(k, "agg_") {
			outField := strings.TrimPrefix(k, "agg_")
			if outField != "" {
				p.Aggregates[outField] = v
			}
		}
	}
	return nil
}

// Process groups the data and computes aggregates.
func (p *AggregateProcessor) Process(ctx context.Context, data []map[string]interface{}) ([]map[string]interface{}, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(data) == 0 {
		return p.BaseProcessor.Process(ctx, data)
	}

	// Grouping
	groups := make(map[string][]map[string]interface{})

	for _, record := range data {
		key := p.generateGroupKey(record)
		groups[key] = append(groups[key], record)
	}

	// Processing groups
	var results []map[string]interface{}

	for _, groupRows := range groups {
		result := make(map[string]interface{})

		// Restore group keys
		// Note: The key generation is lossy for strict restoration if generic strings used,
		// but we can trust the first row of the group for the grouping values.
		if len(groupRows) > 0 {
			first := groupRows[0]
			for _, k := range p.GroupBy {
				result[k] = first[k]
			}
		}

		// Calculate aggregates
		for outField, opSpec := range p.Aggregates {
			parts := strings.SplitN(opSpec, ":", 2)
			op := strings.ToLower(parts[0])
			field := ""
			if len(parts) > 1 {
				field = parts[1]
			}

			val, err := p.calculate(op, field, groupRows)
			if err != nil {
				// We log or error? For now, error out to be safe.
				return nil, fmt.Errorf("aggregate processor: calc failed for %s: %w", outField, err)
			}
			result[outField] = val
		}

		results = append(results, result)
	}

	// Deterministic order for tests
	sort.Slice(results, func(i, j int) bool {
		// simple sort by first group key if available
		if len(p.GroupBy) > 0 {
			k := p.GroupBy[0]
			return fmt.Sprintf("%v", results[i][k]) < fmt.Sprintf("%v", results[j][k])
		}
		return false
	})

	return p.BaseProcessor.Process(ctx, results)
}

func (p *AggregateProcessor) generateGroupKey(record map[string]interface{}) string {
	var sb strings.Builder
	for _, k := range p.GroupBy {
		sb.WriteString(fmt.Sprintf("%v|", record[k]))
	}
	return sb.String()
}

func (p *AggregateProcessor) calculate(op, field string, rows []map[string]interface{}) (interface{}, error) {
	switch op {
	case "count":
		return len(rows), nil
	case "sum", "avg":
		sum := 0.0
		count := 0
		for _, row := range rows {
			val, ok := row[field]
			if !ok || val == nil {
				continue
			}
			f, err := toFloat(val)
			if err != nil {
				continue // skip non-numeric
			}
			sum += f
			count++
		}
		if op == "sum" {
			return sum, nil
		}
		if count == 0 {
			return 0.0, nil
		}
		return sum / float64(count), nil
	case "min", "max":
		var current float64
		set := false
		for _, row := range rows {
			val, ok := row[field]
			if !ok || val == nil {
				continue
			}
			f, err := toFloat(val)
			if err != nil {
				continue
			}
			if !set {
				current = f
				set = true
			} else {
				if op == "min" && f < current {
					current = f
				} else if op == "max" && f > current {
					current = f
				}
			}
		}
		if !set {
			return nil, nil // null if no data
		}
		return current, nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
}

func toFloat(v interface{}) (float64, error) {
	switch i := v.(type) {
	case int:
		return float64(i), nil
	case float64:
		return i, nil
	case int64:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", v)
	}
}
