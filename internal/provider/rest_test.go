package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRESTProvider_Fetch(t *testing.T) {
	tests := []struct {
		name         string
		responseBody interface{}
		statusCode   int
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "array response",
			responseBody: []map[string]interface{}{{"id": 1}, {"id": 2}},
			statusCode:   http.StatusOK,
			wantCount:    2,
			wantErr:      false,
		},
		{
			name:         "single object response",
			responseBody: map[string]interface{}{"id": 1},
			statusCode:   http.StatusOK,
			wantCount:    1,
			wantErr:      false,
		},
		{
			name:         "error status",
			responseBody: map[string]string{"error": "not found"},
			statusCode:   http.StatusNotFound,
			wantCount:    0,
			wantErr:      true,
		},
		{
			name:         "invalid json",
			responseBody: "invalid-json", // This will be encoded as string, so valid JSON string, but invalid structure for our logic if we expect list/map
			statusCode:   http.StatusOK,
			wantCount:    0,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer ts.Close()

			p := NewRESTProvider()
			p.URL = ts.URL

			// Test with default client
			results, err := p.Fetch(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("RESTProvider.Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(results) != tt.wantCount {
					t.Errorf("RESTProvider.Fetch() count = %d, want %d", len(results), tt.wantCount)
				}
			}
		})
	}
}

func TestRESTProvider_Configure(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{
			name: "valid config",
			params: map[string]string{
				"url":         "http://api.example.com",
				"method":      "POST",
				"header_Auth": "Bearer token",
				"timeout":     "5s",
			},
			wantErr: false,
		},
		{
			name:    "missing url",
			params:  map[string]string{"method": "GET"},
			wantErr: true,
		},
		{
			name:    "invalid timeout",
			params:  map[string]string{"url": "http://a.com", "timeout": "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewRESTProvider()
			if err := p.Configure(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("RESTProvider.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.params["header_Auth"] != "" {
				if p.Headers["Auth"] != "Bearer token" {
					t.Errorf("Header not set correctly, got %v", p.Headers["Auth"])
				}
			}
		})
	}
}

func TestRESTProvider_Headers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "MyValue" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	p := NewRESTProvider()
	p.Configure(map[string]string{
		"url":             ts.URL,
		"header_X-Custom": "MyValue",
	})

	_, err := p.Fetch(context.Background())
	if err != nil {
		t.Errorf("Fetch failed with headers: %v", err)
	}
}

func TestRESTProvider_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	p := NewRESTProvider()
	p.URL = ts.URL
	p.Timeout = 1 * time.Millisecond // very short timeout

	_, err := p.Fetch(context.Background())
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
