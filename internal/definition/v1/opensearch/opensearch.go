package opensearch

const (
	Module = "opensearch"
)

type Dashboard struct {
	Link    string `json:"link"`
	Enabled bool   `json:"enabled"`
}
