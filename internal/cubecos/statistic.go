package cubecos

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

type Summary struct {
	Vm      `json:"vm"`
	Role    `json:"role"`
	Metrics definition.Metrics `json:"metrics"`
}

type Vm struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Suspend int `json:"suspend"`
	Paused  int `json:"paused"`
	Error   int `json:"error"`
	Unknown int `json:"unknown"`
}

type Role struct {
	ControlConverged int `json:"controlConverged"`
	Control          int `json:"control"`
	Compute          int `json:"compute"`
	Storage          int `json:"storage"`
	Others           int `json:"others"`
}
