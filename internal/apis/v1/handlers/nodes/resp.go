package nodes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type nodePage struct {
	Nodes      []nodes.Node `json:"nodes"`
	pages.Page `json:"page"`
}
