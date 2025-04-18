package cubecos

import (
	"fmt"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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
	nodes, err := ListNodes()
	if err != nil {
		return nil, err
	}

	role := &Role{}
	for _, node := range nodes {
		switch node.Role {
		case v1.RoleControl:
			role.Control.Count++
		case v1.RoleCompute:
			role.Compute.Count++
		case v1.RoleStorage:
			role.Storage.Count++
		case v1.RoleControlConverged:
			role.ControlConverged.Count++
		case v1.RoleEdgeCore:
			role.EdgeCore.Count++
		case v1.RoleModerator:
			role.Moderator.Count++
		}
	}

	return role, nil
}
