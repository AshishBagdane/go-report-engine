package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLProvider_Fetch(t *testing.T) {
	// Create mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	query := "SELECT id, name FROM users"

	// Success case
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectPing()
	mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)

	p := &SQLProvider{
		Query: query,
		db:    db, // Inject mock DB
	}

	results, err := p.Fetch(context.Background())
	if err != nil {
		t.Errorf("SQLProvider.Fetch() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if name := results[0]["name"]; name != "Alice" {
		t.Errorf("Expected first name Alice, got %v", name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSQLProvider_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name: "valid config",
			params: map[string]string{
				"driver": "postgres",
				"dsn":    "postgres://user:pass@localhost:5432/db",
				"query":  "SELECT * FROM table",
			},
			wantErr: false,
		},
		{
			name: "missing driver",
			params: map[string]string{
				"dsn":   "dsn",
				"query": "query",
			},
			wantErr: true,
		},
		{
			name: "missing dsn",
			params: map[string]string{
				"driver": "postgres",
				"query":  "query",
			},
			wantErr: true,
		},
		{
			name: "missing query",
			params: map[string]string{
				"driver": "postgres",
				"dsn":    "dsn",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewSQLProvider()
			if err := p.Configure(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("SQLProvider.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLProvider_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectPing()
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db error"))

	p := &SQLProvider{
		Query: "SELECT * FROM users",
		db:    db,
	}

	_, err = p.Fetch(context.Background())
	if err == nil {
		t.Error("Expected error from Fetch(), got nil")
	}
}
