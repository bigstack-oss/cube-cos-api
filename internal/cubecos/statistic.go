package cubecos

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	json "github.com/json-iterator/go"
)

type Summary struct {
	DataCenter DataCenterSummary `json:"dataCenter"`
	Host       HostSummary       `json:"host"`
	Vm         VmSummary         `json:"vm"`
}

type DataCenterSummary struct {
	Usage metric.DataCenterUsage `json:"usage"`
}

type HostSummary struct {
	Role   `json:"role"`
	Usages []HostUsage `json:"usages"`
}

type HostUsage struct {
	Role    string         `json:"role"`
	Name    string         `json:"name"`
	Address string         `json:"address"`
	Cpu     metric.Compute `json:"cpu"`
	Memory  metric.Space   `json:"memory"`
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
	Count  int            `json:"count"`
	Cpu    metric.Compute `json:"cpu"`
	Memory metric.Space   `json:"memory"`
}

type VmSummary struct {
	Status         VmStatus `json:"status"`
	metric.VmUsage `json:"usage"`
}

type VmStatus struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Suspend int `json:"suspend"`
	Paused  int `json:"paused"`
	Error   int `json:"error"`
}

func (h *HostSummary) ListCpuUsages() []metric.Compute {
	var list []metric.Compute
	for _, u := range h.Usages {
		list = append(list, u.Cpu)
	}

	return list
}

func (h *HostSummary) ListMemoryUsages() []metric.Space {
	var list []metric.Space
	for _, u := range h.Usages {
		list = append(list, u.Memory)
	}

	return list
}

func (h *HostSummary) SetHostUsages(nodes []nodes.Node) {
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

func (h *HostSummary) SetRoleUsages() {
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

func (h *HostSummary) setRoleUsage(role map[string]RoleUsage) {
	h.Role.ControlConverged = role[nodes.RoleControlConverged]
	h.Role.Control = role[nodes.RoleControl]
	h.Role.Compute = role[nodes.RoleCompute]
	h.Role.Storage = role[nodes.RoleStorage]
	h.Role.EdgeCore = role[nodes.RoleEdgeCore]
	h.Role.Moderator = role[nodes.RoleModerator]
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
