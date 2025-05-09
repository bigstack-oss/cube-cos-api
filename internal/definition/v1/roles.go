package v1

// const (
// 	RoleControl          = "control"
// 	RoleCompute          = "compute"
// 	RoleStorage          = "storage"
// 	RoleControlConverged = "control-converged"
// 	RoleModerator        = "moderator"
// 	RoleEdgeCore         = "edge-core"
// )

// var (
// 	CurrentRole     string
// 	Roles           = []string{RoleControlConverged, RoleControl, RoleCompute, RoleStorage, RoleModerator, RoleEdgeCore}
// 	updateRoleNodes = sync.Mutex{}

// 	ControlRole          = newControlRole()
// 	ComputeRole          = newComputeRole()
// 	StorageRole          = newStorageRole()
// 	ControlConvergedRole = newControlConvergedRole()
// 	ModeratorRole        = newModeratorRole()
// 	EdgeCoreRole         = newEdgeCoreRole()

// 	AllRoles = []*Role{
// 		ControlRole,
// 		ComputeRole,
// 		StorageRole,
// 		ControlConvergedRole,
// 		ModeratorRole,
// 		EdgeCoreRole,
// 	}

// 	AllGeneralRoles = []*Role{
// 		ControlRole,
// 		ComputeRole,
// 		StorageRole,
// 		ControlConvergedRole,
// 	}

// 	ControlRoles = []*Role{
// 		ControlRole,
// 		ControlConvergedRole,
// 	}

// 	ComputeRoles = []*Role{
// 		ComputeRole,
// 		ControlConvergedRole,
// 		EdgeCoreRole,
// 	}
// )

// type Role struct {
// 	Name  string `json:"name" bson:"name"`
// 	Hosts []Host `json:"hosts" bson:"hosts"`
// 	Nodes []Node `json:"-"`
// }

// type Host struct {
// 	Role string `json:"role,omitzero"`
// 	Name string `json:"name"`
// 	Ip   string `json:"ip,omitzero"`
// }

// func (h *Host) GetNode() *Node {
// 	node, err := GetNodeByHostname(h.Name)
// 	if err != nil {
// 		return nil
// 	}

// 	return node
// }

// func newControlRole() *Role {
// 	return &Role{Name: RoleControl}
// }

// func newComputeRole() *Role {
// 	return &Role{Name: RoleCompute}
// }

// func newStorageRole() *Role {
// 	return &Role{Name: RoleStorage}
// }

// func newControlConvergedRole() *Role {
// 	return &Role{Name: RoleControlConverged}
// }

// func newModeratorRole() *Role {
// 	return &Role{Name: RoleModerator}
// }

// func newEdgeCoreRole() *Role {
// 	return &Role{Name: RoleEdgeCore}
// }

// func GetControlRole() *Role {
// 	return ControlRole
// }

// func GetControlRoles() []*Role {
// 	return ControlRoles
// }

// func GetComputeRole() *Role {
// 	return ComputeRole
// }

// func GetStorageRole() *Role {
// 	return StorageRole
// }

// func GetControlConvergeRole() *Role {
// 	return ControlConvergedRole
// }

// func GetModeratorRole() *Role {
// 	return ModeratorRole
// }

// func GetEdgeCoreRole() *Role {
// 	return EdgeCoreRole
// }

// func SyncRoleNodes() {
// 	updateRoleNodes.Lock()
// 	defer updateRoleNodes.Unlock()

// 	for _, role := range Roles {
// 		nodes, err := GetNodesByRole(role)
// 		if err != nil {
// 			return
// 		}

// 		role := GetRole(role)
// 		if role != nil {
// 			role.Nodes = nodes
// 			role.Hosts = convertNodesToHosts(nodes)
// 		}
// 	}
// }

// func convertNodesToHosts(nodes []Node) []Host {
// 	hosts := []Host{}
// 	for _, node := range nodes {
// 		hosts = append(hosts, Host{
// 			Name: node.Hostname,
// 			Ip:   nodes.Ip,
// 		})
// 	}

// 	return hosts
// }

// func parseNodes(svc *registry.Service) []Node {
// 	nodes := []Node{}
// 	for _, node := range svc.Nodes {
// 		nodes = append(nodes, newNode(node))
// 	}

// 	return nodes
// }

// func parseNodesByRole(svc *registry.Service, roleName string) []Node {
// 	nodes := []Node{}
// 	for _, node := range svc.Nodes {
// 		if nodes.Metadata["role"] != roleName {
// 			continue
// 		}

// 		nodes = append(nodes, newNode(node))
// 	}

// 	return nodes
// }

// func newNode(node *registry.Node) Node {
// 	return Node{
// 		Role:         nodes.Metadata["role"],
// 		Id:           nodes.Metadata["nodeID"],
// 		SerialNumber: nodes.Metadata["serialNumber"],
// 		DataCenter:   nodes.Metadata["dataCenter"],
// 		Protocol:     nodes.Metadata["protocol"],
// 		Hostname:     nodes.Metadata["hostname"],
// 		Ip:           nodes.Metadata["ip"],
// 		Address:      nodes.Address,
// 		Labels: map[string]string{
// 			"isGpuEnabled": nodes.Metadata["isGpuEnabled"],
// 		},
// 	}
// }

// func GetRole(name string) *Role {
// 	switch name {
// 	case RoleControl:
// 		return ControlRole
// 	case RoleCompute:
// 		return ComputeRole
// 	case RoleStorage:
// 		return StorageRole
// 	case RoleControlConverged:
// 		return ControlConvergedRole
// 	case RoleModerator:
// 		return ModeratorRole
// 	case RoleEdgeCore:
// 		return EdgeCoreRole
// 	}

// 	return nil
// }

// func (r *Role) IsNodeEmpty() bool {
// 	return len(r.Nodes) == 0
// }
