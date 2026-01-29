package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RESTProvider implements ProviderStrategy for fetching data from REST APIs.
type RESTProvider struct {
	URL     string
	Method  string
	Headers map[string]string
	Timeout time.Duration

	// client allows injection of custom http client (useful for mocks or specific configs)
	client *http.Client
}

// NewRESTProvider creates a new instance of RESTProvider with defaults.
func NewRESTProvider() *RESTProvider {
	return &RESTProvider{
		Method:  "GET",
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}
}

// Fetch calls the REST API and processes the JSON response.
func (p *RESTProvider) Fetch(ctx context.Context) ([]map[string]interface{}, error) {
	if p.URL == "" {
		return nil, fmt.Errorf("rest provider: url not configured")
	}

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("rest provider: failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}

	// Use custom client or default
	client := p.client
	if client == nil {
		client = &http.Client{
			Timeout: p.Timeout,
		}
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rest provider: request failed: %w", err)
	}
	defer resp.Body.Close()

	// Validate status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("rest provider: api returned error status: %d %s", resp.StatusCode, resp.Status)
	}

	// Decode response
	// We try to decode as []map[string]interface{}.
	// If that fails, we try map[string]interface{} and wrap it.
	var raw interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("rest provider: failed to decode json body: %w", err)
	}

	// Normalize data
	var results []map[string]interface{}

	switch v := raw.(type) {
	case []interface{}:
		// Array of objects
		for i, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				results = append(results, m)
			} else {
				// Warn or skip? skipping non-object items in array
				// Ideally logging would happen here but we don't have logger injected yet in this simple struct.
				return nil, fmt.Errorf("rest provider: item at index %d is not a json object", i)
			}
		}
	case map[string]interface{}:
		// Single object
		results = append(results, v)
	default:
		return nil, fmt.Errorf("rest provider: response must be a json object or array of objects")
	}

	return results, nil
}

// Configure sets up the provider from a map of parameters.
// Params:
// - url: API URL (required)
// - method: HTTP Method (default: "GET")
// - header_<KEY>: Custom headers, e.g., "header_Authorization"
// - timeout: Timeout duration string (e.g., "30s", "1m")
func (p *RESTProvider) Configure(params map[string]string) error {
	if url, ok := params["url"]; ok {
		p.URL = url
	} else {
		return fmt.Errorf("rest provider: missing required parameter 'url'")
	}

	if method, ok := params["method"]; ok {
		p.Method = strings.ToUpper(method)
	}

	if timeoutStr, ok := params["timeout"]; ok {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("rest provider: invalid timeout %s: %w", timeoutStr, err)
		}
		p.Timeout = d
	}

	// Parse headers (prefix "header_")
	for k, v := range params {
		if strings.HasPrefix(k, "header_") {
			key := strings.TrimPrefix(k, "header_")
			if key != "" {
				p.Headers[key] = v
			}
		}
	}

	return nil
}
