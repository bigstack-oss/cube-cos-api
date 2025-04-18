package cubecos

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	json "github.com/json-iterator/go"
)

type Summary struct {
	DataCenter DataCenterSummary `json:"dataCenter"`
	Host       HostSummary       `json:"host"`
	Vm         VmSummary         `json:"vm"`
}

type DataCenterSummary struct {
	Usage v1.DataCenterUsage `json:"usage"`
}

type HostSummary struct {
	Role   `json:"role"`
	Usages []HostUsage `json:"usages"`
}

func (h *HostSummary) ListCpuUsages() []v1.ComputeStatistic {
	var list []v1.ComputeStatistic
	for _, u := range h.Usages {
		list = append(list, u.Cpu)
	}

	return list
}

func (h *HostSummary) ListMemoryUsages() []v1.SpaceStatistic {
	var list []v1.SpaceStatistic
	for _, u := range h.Usages {
		list = append(list, u.Memory)
	}

	return list
}

func (h *HostSummary) SetHostUsageByNodes(nodes []v1.Node) {
	for _, node := range nodes {
		usage, err := GetHostUsage(node)
		if err != nil {
			continue
		}

		h.Usages = append(
			h.Usages,
			HostUsage{
				Role:    node.Role,
				Name:    node.Hostname,
				Address: node.Ip,
				Cpu:     usage.Cpu,
				Memory:  usage.Memory,
			},
		)
	}
}

func (h *HostSummary) SetRoleUsageByHosts() {
	roleMap := map[string]RoleUsage{}
	h.sumRoleUsage(roleMap)
	h.summarizeRoleUsage(roleMap)
	h.setRoleUsage(roleMap)
}

func (h *HostSummary) sumRoleUsage(roleMap map[string]RoleUsage) {
	for _, u := range h.Usages {
		role, found := roleMap[u.Role]
		if !found {
			role = RoleUsage{}
		}

		role.Count++
		role.Cpu.TotalCores += u.Cpu.TotalCores
		role.Cpu.UsedCores += u.Cpu.UsedCores
		role.Cpu.UsedPercent += u.Cpu.UsedPercent
		role.Cpu.FreeCores += u.Cpu.FreeCores
		role.Cpu.FreePercent += u.Cpu.FreePercent

		role.Memory.TotalMiB += u.Memory.TotalMiB
		role.Memory.UsedMiB += u.Memory.UsedMiB
		role.Memory.UsedPercent += u.Memory.UsedPercent
		role.Memory.FreeMiB += u.Memory.FreeMiB
		role.Memory.FreePercent += u.Memory.FreePercent

		roleMap[u.Role] = role
	}
}

func (h *HostSummary) summarizeRoleUsage(roleMap map[string]RoleUsage) {
	for role, usage := range roleMap {
		usage.Cpu.TotalCores = math.RoundDown(usage.Cpu.TotalCores, 4)
		usage.Cpu.UsedCores = math.RoundDown(usage.Cpu.UsedCores, 4)
		usage.Cpu.FreeCores = math.RoundDown(usage.Cpu.FreeCores, 4)
		usage.Cpu.UsedPercent = math.RoundDown(usage.Cpu.UsedPercent/float64(usage.Count), 4)
		usage.Cpu.FreePercent = math.RoundDown(usage.Cpu.FreePercent/float64(usage.Count), 4)
		usage.Memory.TotalMiB = math.RoundDown(usage.Memory.TotalMiB, 4)
		usage.Memory.UsedMiB = math.RoundDown(usage.Memory.UsedMiB, 4)
		usage.Memory.FreeMiB = math.RoundDown(usage.Memory.FreeMiB, 4)
		usage.Memory.UsedPercent = math.RoundDown(usage.Memory.UsedPercent/float64(usage.Count), 4)
		usage.Memory.FreePercent = math.RoundDown(usage.Memory.FreePercent/float64(usage.Count), 4)
		roleMap[role] = usage
	}
}

func (h *HostSummary) setRoleUsage(roleMap map[string]RoleUsage) {
	h.Role.ControlConverged = roleMap[v1.RoleControlConverged]
	h.Role.Control = roleMap[v1.RoleControl]
	h.Role.Compute = roleMap[v1.RoleCompute]
	h.Role.Storage = roleMap[v1.RoleStorage]
	h.Role.EdgeCore = roleMap[v1.RoleEdgeCore]
	h.Role.Moderator = roleMap[v1.RoleModerator]
}

type HostUsage struct {
	Role    string              `json:"role"`
	Name    string              `json:"name"`
	Address string              `json:"address"`
	Cpu     v1.ComputeStatistic `json:"cpu"`
	Memory  v1.SpaceStatistic   `json:"memory"`
}

type Role struct {
	ControlConverged RoleUsage `json:"controlConverged"`
	Control          RoleUsage `json:"control"`
	Compute          RoleUsage `json:"compute"`
	Storage          RoleUsage `json:"storage"`
	EdgeCore         RoleUsage `json:"edgeCore"`
	Moderator        RoleUsage `json:"moderator"`
}

type RoleUsage struct {
	Count  int                 `json:"count"`
	Cpu    v1.ComputeStatistic `json:"cpu"`
	Memory v1.SpaceStatistic   `json:"memory"`
}

type VmSummary struct {
	Status     VmStatus `json:"status"`
	v1.VmUsage `json:"usage"`
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

func (s *Summary) String() string {
	return string(s.Bytes())
}
