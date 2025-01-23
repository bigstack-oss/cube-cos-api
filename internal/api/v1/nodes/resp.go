package nodes

import definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"

type data struct {
	Nodes           []*definition.Node `json:"nodes"`
	definition.Page `json:"page"`
}
