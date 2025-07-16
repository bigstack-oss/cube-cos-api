package nodes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
)

type nodePage struct {
	Nodes      []nodes.Node `json:"nodes"`
	pages.Page `json:"page"`
}

type devicesResp struct {
	Code   int                 `json:"code"`
	Status string              `json:"status"`
	Msg    string              `json:"msg"`
	Data   []nodes.BlockDevice `json:"data"`
}
