package v1

import (
	"sync"

	"go-micro.dev/v5/registry"
)

const (
	RoleControl          = "control"
	RoleCompute          = "compute"
	RoleStorage          = "storage"
	RoleNetwork          = "network"
	RoleControlConverged = "control-converged"
	RoleModerator        = "moderator"
	RoleEdgeCore         = "edge-core"
)

var (
	CurrentRole string
	Roles       = []string{RoleControl, RoleCompute, RoleStorage, RoleNetwork, RoleControlConverged, RoleModerator, RoleEdgeCore}
	update      = sync.Mutex{}

	ControlRole          = newControlRole()
	ComputeRole          = newComputeRole()
	StorageRole          = newStorageRole()
	NetworkRole          = newNetworkRole()
	ControlConvergedRole = newControlConvergeRole()
	ModeratorRole        = newModeratorRole()
	EdgeCoreRole         = newEdgeCoreRole()

	AllRoles = []*Role{
		ControlRole,
		ComputeRole,
		StorageRole,
		NetworkRole,
		ControlConvergedRole,
		ModeratorRole,
		EdgeCoreRole,
	}

	AllGeneralRoles = []*Role{
		ControlRole,
		ComputeRole,
		StorageRole,
		NetworkRole,
		ControlConvergedRole,
	}

	ControlRoles = []*Role{
		ControlRole,
		ControlConvergedRole,
	}

	ComputeRoles = []*Role{
		ComputeRole,
		ControlConvergedRole,
		EdgeCoreRole,
	}
)

type Role struct {
	Name  string  `json:"name" bson:"name"`
	Nodes []*Node `json:"nodes" bson:"nodes"`
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

func newNetworkRole() *Role {
	return &Role{Name: RoleNetwork}
}

func newControlConvergeRole() *Role {
	return &Role{Name: RoleControlConverged}
}

func newModeratorRole() *Role {
	return &Role{Name: RoleModerator}
}

func newEdgeCoreRole() *Role {
	return &Role{Name: RoleEdgeCore}
}

func GetControlRole() *Role {
	return ControlRole
}

func GetControlRoles() []*Role {
	return ControlRoles
}

func GetComputeRole() *Role {
	return ComputeRole
}

func GetStorageRole() *Role {
	return StorageRole
}

func GetNetworkRole() *Role {
	return NetworkRole
}

func GetControlConvergeRole() *Role {
	return ControlConvergedRole
}

func GetModeratorRole() *Role {
	return ModeratorRole
}

func GetEdgeCoreRole() *Role {
	return EdgeCoreRole
}

func SyncNodesOfRole() {
	update.Lock()
	defer update.Unlock()

	for _, role := range Roles {
		newNodes, err := GetNodesByRole(role)
		if err != nil {
			return
		}

		role := getRole(role)
		if role != nil {
			role.Nodes = newNodes
		}
	}
}

func getNodesByService(svc *registry.Service, role string) []*Node {
	nodes := []*Node{}

	for _, node := range svc.Nodes {
		nodes = append(
			nodes,
			genNodeInfo(node, role),
		)
	}

	return nodes
}

func genNodeInfo(node *registry.Node, role string) *Node {
	return &Node{
		Role:     role,
		ID:       node.Metadata["nodeID"],
		Hostname: node.Metadata["hostname"],
		Address:  node.Address,
		Labels: map[string]string{
			"isGPUEnabled": node.Metadata["isGPUEnabled"],
		},
	}
}

func getRole(name string) *Role {
	switch name {
	case RoleControl:
		return ControlRole
	case RoleCompute:
		return ComputeRole
	case RoleStorage:
		return StorageRole
	case RoleNetwork:
		return NetworkRole
	case RoleControlConverged:
		return ControlConvergedRole
	case RoleModerator:
		return ModeratorRole
	case RoleEdgeCore:
		return EdgeCoreRole
	}

	return nil
}

func (r *Role) IsNodeEmpty() bool {
	return r.Nodes == nil || len(r.Nodes) == 0
}
