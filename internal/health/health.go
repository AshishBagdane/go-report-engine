package health

import "context"

// Status represents the health status of a component.
type Status string

const (
	StatusUp       Status = "UP"
	StatusDown     Status = "DOWN"
	StatusDegraded Status = "DEGRADED"
)

// Result represents the result of a health check.
type Result struct {
	Status  Status                 `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// Checker is the interface implemented by components that support health checks.
type Checker interface {
	CheckHealth(ctx context.Context) (Result, error)
}
