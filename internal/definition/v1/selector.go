package v1

type Selector struct {
	Enabled bool              `json:"enabled" yaml:"enabled"`
	Labels  map[string]string `json:"labels" yaml:"labels"`
}

type Label struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}
