package provider

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLProvider implements ProviderStrategy for fetching data from a SQL database.
// It uses the standard database/sql interface, so it supports any driver (postgres, mysql, sqlite, etc.).
// Note: The driver MUST be imported in the main application (e.g. _ "github.com/lib/pq").
type SQLProvider struct {
	Driver string
	DSN    string
	Query  string

	// db, if set allows reusing an existing connection pool,
	// otherwise one is created per Fetch (and closed).
	// For production efficiency, managing the DB connection externally is better,
	// but for this simple provider we will open/close if not provided.
	db *sql.DB
}

// NewSQLProvider creates a new generic SQL provider.
func NewSQLProvider() *SQLProvider {
	return &SQLProvider{}
}

// Fetch executes the configured query and returns the results.
func (p *SQLProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	if p.Query == "" {
		return nil, fmt.Errorf("sql provider: query not configured")
	}

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var db *sql.DB
	var err error
	var shouldClose bool

	if p.db != nil {
		db = p.db
	} else {
		if p.Driver == "" || p.DSN == "" {
			return nil, fmt.Errorf("sql provider: driver and dsn are required if external db not provided")
		}
		db, err = sql.Open(p.Driver, p.DSN)
		if err != nil {
			return nil, fmt.Errorf("sql provider: failed to open connection: %w", err)
		}
		shouldClose = true
	}

	if shouldClose {
		defer func() { _ = db.Close() }()
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sql provider: ping failed: %w", err)
	}

	// Execute query
	rows, err := db.QueryContext(ctx, p.Query)
	if err != nil {
		return nil, fmt.Errorf("sql provider: query failed: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// Get columns
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sql provider: failed to get columns: %w", err)
	}

	// Prepare results
	var results []map[string]interface{}

	// Helper for scanning
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)

	for rows.Next() {
		// Prepare pointers for scanning
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("sql provider: scan failed: %w", err)
		}

		// Map to result
		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle bytes -> string conversion for common SQL drivers
			// Some drivers return []byte for strings/text
			if b, ok := val.([]byte); ok {
				entry[col] = string(b)
			} else {
				entry[col] = val
			}
		}

		results = append(results, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sql provider: row iteration error: %w", err)
	}

	return results, nil
}

// Configure sets up the provider from a map of parameters.
// Params:
// - driver: Database driver name (e.g., "postgres", "mysql"). Required.
// - dsn: Data Source Name connection string. Required.
// - query: SQL Query to execute. Required.
func (p *SQLProvider) Configure(params map[string]string) error {
	if driver, ok := params["driver"]; ok {
		p.Driver = driver
	} else {
		return fmt.Errorf("sql provider: missing required parameter 'driver'")
	}

	if dsn, ok := params["dsn"]; ok {
		p.DSN = dsn
	} else {
		return fmt.Errorf("sql provider: missing required parameter 'dsn'")
	}

	if query, ok := params["query"]; ok {
		p.Query = query
	} else {
		return fmt.Errorf("sql provider: missing required parameter 'query'")
	}

	return nil
}

// Close allows cleaning up if we managed the connection logic differently,
// but here Fetch manages individual connections or uses a shared pool.
// Implementing Closeable logic would be needed if we kept a long-lived connection.
// For now, this is stateless/ephemeral per Fetch or uses external DB.
