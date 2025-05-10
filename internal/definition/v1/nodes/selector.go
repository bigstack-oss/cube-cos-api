package nodes

type Selector struct {
	Enabled bool              `json:"enabled" yaml:"enabled"`
	Labels  map[string]string `json:"labels" yaml:"labels"`
}
