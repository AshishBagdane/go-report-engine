package formatter

type FormatStrategy interface {
	Format(data []map[string]interface{}) ([]byte, error)
}
