package nodes

import (
	"slices"
	"sync/atomic"
)

const (
	RoleControl          = "control"
	RoleCompute          = "compute"
	RoleStorage          = "storage"
	RoleControlConverged = "control-converged"
	RoleModerator        = "moderator"
	RoleEdgeCore         = "edge-core"
)

var (
	roles = []string{
		RoleControlConverged,
		RoleControl,
		RoleCompute,
		RoleStorage,
		RoleModerator,
		RoleEdgeCore,
	}

	Control          atomic.Pointer[Role]
	Compute          atomic.Pointer[Role]
	Storage          atomic.Pointer[Role]
	ControlConverged atomic.Pointer[Role]
	Moderator        atomic.Pointer[Role]
	EdgeCore         atomic.Pointer[Role]

	AllRoles = []*Role{
		Control.Load(),
		Compute.Load(),
		Storage.Load(),
		ControlConverged.Load(),
		Moderator.Load(),
		EdgeCore.Load(),
	}

	AllGeneralRoles = []*Role{
		Control.Load(),
		Compute.Load(),
		Storage.Load(),
		ControlConverged.Load(),
	}

	ControlRoles = []*Role{
		Control.Load(),
		ControlConverged.Load(),
	}

	ComputeRoles = []*Role{
		Compute.Load(),
		ControlConverged.Load(),
		EdgeCore.Load(),
	}

	cloudRoles = []string{
		RoleControlConverged,
		RoleControl,
		RoleCompute,
		RoleStorage,
	}
	edgeRoles = []string{
		RoleEdgeCore,
		RoleModerator,
	}
)

func init() {
	newControlRole()
	newComputeRole()
	newStorageRole()
	newControlConvergedRole()
	newModeratorRole()
	newEdgeCoreRole()
}

type Role struct {
	Name  string `json:"name" bson:"name"`
	Hosts []Host `json:"hosts" bson:"hosts"`
	Nodes []Node `json:"-"`
}

type Host struct {
	Role string `json:"role,omitzero"`
	Name string `json:"name"`
	Ip   string `json:"ip,omitzero"`
}

func (r *Role) IsNodeEmpty() bool {
	return len(r.Nodes) == 0
}

func (h *Host) GetNode() *Node {
	node, err := Get(h.Name)
	if err != nil {
		return nil
	}

	return node
}

func newControlRole() {
	Control.Store(&Role{Name: RoleControl})
}

func newComputeRole() {
	Compute.Store(&Role{Name: RoleCompute})
}

func newStorageRole() {
	Storage.Store(&Role{Name: RoleStorage})
}

func newControlConvergedRole() {
	ControlConverged.Store(&Role{Name: RoleControlConverged})
}

func newModeratorRole() {
	Moderator.Store(&Role{Name: RoleModerator})
}

func newEdgeCoreRole() {
	EdgeCore.Store(&Role{Name: RoleEdgeCore})
}

func GetControlRole() *Role {
	return Control.Load()
}

func GetComputeRole() *Role {
	return Compute.Load()
}

func GetStorageRole() *Role {
	return Storage.Load()
}

func GetControlConvergeRole() *Role {
	return ControlConverged.Load()
}

func GetModeratorRole() *Role {
	return Moderator.Load()
}

func GetEdgeCoreRole() *Role {
	return EdgeCore.Load()
}

func GetCloudRoles() []string {
	return cloudRoles
}

func GetEdgeRoles() []string {
	return edgeRoles
}

func IsCloudRole(role string) bool {
	return slices.Contains(cloudRoles, role)
}

func IsEdgeRole(role string) bool {
	return slices.Contains(edgeRoles, role)
}
