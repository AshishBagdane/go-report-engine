package provider

type ProviderStrategy interface {
	Fetch() ([]map[string]interface{}, error)
}
