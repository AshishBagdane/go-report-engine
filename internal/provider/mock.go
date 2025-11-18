package provider

func NewMockProvider() ProviderStrategy {
	return &MockProvider{}
}

type MockProvider struct{}

func (m *MockProvider) Fetch() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"id":    1,
			"name":  "Alice",
			"score": 95,
		},
		{
			"id":    2,
			"name":  "Bob",
			"score": 88,
		},
	}, nil
}
