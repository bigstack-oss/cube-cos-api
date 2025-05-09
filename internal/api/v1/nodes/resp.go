package nodes

import (
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
)

type nodePage struct {
	Nodes   []nodes.Node `json:"nodes"`
	v1.Page `json:"page"`
}
