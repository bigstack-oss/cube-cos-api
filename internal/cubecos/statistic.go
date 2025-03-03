package cubecos

import (
	json "github.com/json-iterator/go"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

type Summary struct {
	DataCenter DataCenterSummary `json:"dataCenter"`
	Host       HostSummary       `json:"host"`
	Vm         VmSummary         `json:"vm"`
}

type DataCenterSummary struct {
	Usage definition.DataCenterUsage `json:"usage"`
}

type HostSummary struct {
	Role   `json:"role"`
	Usages []HostUsage `json:"usages"`
}

func (h *HostSummary) ListCpuUsages() []definition.ComputeStatistic {
	var list []definition.ComputeStatistic
	for _, u := range h.Usages {
		list = append(list, u.Cpu)
	}

	return list
}

func (h *HostSummary) ListMemoryUsages() []definition.SpaceStatistic {
	var list []definition.SpaceStatistic
	for _, u := range h.Usages {
		list = append(list, u.Memory)
	}

	return list
}

type HostUsage struct {
	Role                 string `json:"role"`
	Name                 string `json:"name"`
	Address              string `json:"address"`
	definition.HostUsage `json:"usage"`
}

type Role struct {
	ControlConverged int `json:"controlConverged"`
	Control          int `json:"control"`
	Compute          int `json:"compute"`
	Storage          int `json:"storage"`
}

type VmSummary struct {
	Status             VmStatus `json:"status"`
	definition.VmUsage `json:"usage"`
}

type VmStatus struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Suspend int `json:"suspend"`
	Paused  int `json:"paused"`
	Error   int `json:"error"`
}

func (s *Summary) Bytes() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		return []byte{}
	}

	return b
}
