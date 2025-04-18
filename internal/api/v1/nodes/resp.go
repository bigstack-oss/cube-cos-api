package nodes

import v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Nodes   []v1.Node `json:"nodes"`
	v1.Page `json:"page"`
}
