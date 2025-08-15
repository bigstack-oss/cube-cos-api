package firmwares

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"

type node struct {
	Name           string `json:"name"`
	nodes.Firmware `json:"firmware"`
}
