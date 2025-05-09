package nodes

import "sync"

const (
	RoleControl          = "control"
	RoleCompute          = "compute"
	RoleStorage          = "storage"
	RoleControlConverged = "control-converged"
	RoleModerator        = "moderator"
	RoleEdgeCore         = "edge-core"
)

var (
	syncNodes sync.Mutex

	roles = []string{
		RoleControlConverged,
		RoleControl,
		RoleCompute,
		RoleStorage,
		RoleModerator,
		RoleEdgeCore,
	}

	Control          = newControlRole()
	Compute          = newComputeRole()
	Storage          = newStorageRole()
	ControlConverged = newControlConvergedRole()
	Moderator        = newModeratorRole()
	EdgeCore         = newEdgeCoreRole()

	AllRoles = []*Role{
		Control,
		Compute,
		Storage,
		ControlConverged,
		Moderator,
		EdgeCore,
	}

	AllGeneralRoles = []*Role{
		Control,
		Compute,
		Storage,
		ControlConverged,
	}

	ControlRoles = []*Role{
		Control,
		ControlConverged,
	}

	ComputeRoles = []*Role{
		Compute,
		ControlConverged,
		EdgeCore,
	}
)

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

func GetRole(name string) *Role {
	switch name {
	case RoleControl:
		return Control
	case RoleCompute:
		return Compute
	case RoleStorage:
		return Storage
	case RoleControlConverged:
		return ControlConverged
	case RoleModerator:
		return Moderator
	case RoleEdgeCore:
		return EdgeCore
	default:
		return nil
	}
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

func newControlRole() *Role {
	return &Role{Name: RoleControl}
}

func newComputeRole() *Role {
	return &Role{Name: RoleCompute}
}

func newStorageRole() *Role {
	return &Role{Name: RoleStorage}
}

func newControlConvergedRole() *Role {
	return &Role{Name: RoleControlConverged}
}

func newModeratorRole() *Role {
	return &Role{Name: RoleModerator}
}

func newEdgeCoreRole() *Role {
	return &Role{Name: RoleEdgeCore}
}

func GetControlRole() *Role {
	return Control
}

func GetControlRoles() []*Role {
	return ControlRoles
}

func GetComputeRole() *Role {
	return Compute
}

func GetStorageRole() *Role {
	return Storage
}

func GetControlConvergeRole() *Role {
	return ControlConverged
}

func GetModeratorRole() *Role {
	return Moderator
}

func GetEdgeCoreRole() *Role {
	return EdgeCore
}
