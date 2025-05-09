package cubecos

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
)

const (
	cubeSysRole = "cubesys.role"
)

func GetNodeRole() (string, error) {
	role, err := GetTuningValue(cubeSysRole)
	if err != nil {
		return "", err
	}

	if role == "" {
		return "", fmt.Errorf("role is empty")
	}

	return role, nil
}

func GetRoleStatus() (*Role, error) {
	role := &Role{}
	for _, n := range nodes.List() {
		switch n.Role {
		case nodes.RoleControl:
			role.Control.Count++
		case nodes.RoleCompute:
			role.Compute.Count++
		case nodes.RoleStorage:
			role.Storage.Count++
		case nodes.RoleControlConverged:
			role.ControlConverged.Count++
		case nodes.RoleEdgeCore:
			role.EdgeCore.Count++
		case nodes.RoleModerator:
			role.Moderator.Count++
		}
	}

	return role, nil
}
