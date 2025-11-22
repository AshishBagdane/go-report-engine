package logging

import "context"

// Context key types for type-safe context values
type contextKey int

const (
	requestIDKey contextKey = iota
	correlationIDKey
)

// WithRequestID adds a request ID to the context.
// Request IDs are used to track individual requests through the system.
//
// Example:
//
//	ctx := logging.WithRequestID(context.Background(), "req-abc-123")
//	logger.InfoContext(ctx, "processing request")  // Will include request_id field
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context.
// Returns empty string if no request ID is set.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithCorrelationID adds a correlation ID to the context.
// Correlation IDs are used to track related operations across services.
//
// Example:
//
//	ctx := logging.WithCorrelationID(context.Background(), "corr-xyz-789")
//	logger.InfoContext(ctx, "processing started")  // Will include correlation_id field
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID retrieves the correlation ID from the context.
// Returns empty string if no correlation ID is set.
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}
