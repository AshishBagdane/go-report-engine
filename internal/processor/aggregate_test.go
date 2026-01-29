package processor

import (
	"context"
	"testing"
)

func TestAggregateProcessor_Process(t *testing.T) {
	data := []map[string]interface{}{
		{"dept": "Sales", "amount": 100, "region": "US"},
		{"dept": "Sales", "amount": 200, "region": "US"},
		{"dept": "Sales", "amount": 50, "region": "EU"},
		{"dept": "Eng", "amount": 300, "region": "US"},
	}

	tests := []struct {
		name       string
		groupBy    []string
		aggregates map[string]string
		wantCount  int
		checks     func([]map[string]interface{}) bool
	}{
		{
			name:       "group by dept, sum amount",
			groupBy:    []string{"dept"},
			aggregates: map[string]string{"total": "sum:amount"},
			wantCount:  2, // Sales, Eng
			checks: func(results []map[string]interface{}) bool {
				// Sales: 100+200+50 = 350
				// Eng: 300
				foundSales := false
				for _, r := range results {
					if r["dept"] == "Sales" {
						foundSales = true
						if r["total"] != 350.0 {
							t.Logf("Sales total = %v, want 350", r["total"])
							return false
						}
					}
				}
				return foundSales
			},
		},
		{
			name:       "group by dept and region, count",
			groupBy:    []string{"dept", "region"},
			aggregates: map[string]string{"count": "count:amount"},
			wantCount:  3, // Sales-US, Sales-EU, Eng-US
			checks: func(results []map[string]interface{}) bool {
				// Sales-US: 2
				for _, r := range results {
					if r["dept"] == "Sales" && r["region"] == "US" {
						if r["count"] != 2 {
							return false
						}
					}
				}
				return true
			},
		},
		{
			name:       "avg calculation",
			groupBy:    []string{"dept"},
			aggregates: map[string]string{"avg_amt": "avg:amount"},
			wantCount:  2,
			checks: func(results []map[string]interface{}) bool {
				// Sales avg: 350/3 = 116.666...
				for _, r := range results {
					if r["dept"] == "Sales" {
						val := r["avg_amt"].(float64)
						if val < 116.6 || val > 116.7 {
							return false
						}
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewAggregateProcessor(tt.groupBy, tt.aggregates)
			got, err := p.Process(context.Background(), data)
			if err != nil {
				t.Fatalf("Process() error = %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("Process() count = %d, want %d", len(got), tt.wantCount)
			}
			if !tt.checks(got) {
				t.Error("Checks failed")
			}
		})
	}
}
