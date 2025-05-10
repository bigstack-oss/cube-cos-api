package grafana

const (
	Module = "grafana"
)

type Dashboard struct {
	Link    string `json:"link"`
	Enabled bool   `json:"enabled"`
}
