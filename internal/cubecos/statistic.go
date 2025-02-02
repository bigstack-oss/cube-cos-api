package cubecos

import (
	json "github.com/json-iterator/go"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

type Summary struct {
	Host HostSummary `json:"host"`
	Vm   VmSummary   `json:"vm"`
}

type HostSummary struct {
	Role             `json:"role"`
	definition.Usage `json:"usage"`
}

type Role struct {
	ControlConverged int `json:"controlConverged"`
	Control          int `json:"control"`
	Compute          int `json:"compute"`
	Storage          int `json:"storage"`
	Others           int `json:"others"`
}

type VmSummary struct {
	Status           VmStatus `json:"status"`
	definition.Usage `json:"usage"`
}

type VmStatus struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Suspend int `json:"suspend"`
	Paused  int `json:"paused"`
	Error   int `json:"error"`
	Unknown int `json:"unknown"`
}

func (s *Summary) Bytes() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		return []byte{}
	}

	return b
}
