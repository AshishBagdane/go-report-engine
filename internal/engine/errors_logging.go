package engine

import "fmt"

// Logging configuration errors
var (
	// ErrInvalidLogLevel indicates an invalid log level was specified
	ErrInvalidLogLevel = fmt.Errorf("invalid log level: must be one of: debug, info, warn, warning, error")

	// ErrInvalidLogFormat indicates an invalid log format was specified
	ErrInvalidLogFormat = fmt.Errorf("invalid log format: must be one of: json, text")
)
