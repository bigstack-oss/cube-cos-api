package v1

const (
	Grafana = "grafana"
)

type Dashboard struct {
	Link    string `json:"link"`
	Enabled bool   `json:"enabled"`
}
