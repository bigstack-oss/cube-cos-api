package fixpacks

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

type node struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	UpdatedAt string `json:"updatedAt"`
}

type progress struct {
	Host   string                 `json:"host"`
	Phase  string                 `json:"phase"`
	Status status.FixpackProgress `json:"status"`
}
