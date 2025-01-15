package cubecos

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1/accelerators/devices"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

const (
	cubeSysRole = "cubesys.role"
)

func GetNodeRole() (string, error) {
	role, err := ReadHexTuning(cubeSysRole)
	if err != nil {
		return "", err
	}

	if role == "" {
		return "", fmt.Errorf("role is empty")
	}

	return role, nil
}

func IsGpuEnabled() bool {
	provider, err := openstack.NewProvider(config.Data.Spec.Dependency.Openstack.ConfFile)
	if err != nil {
		log.Errorf("failed to create openstack provider: %s", err.Error())
		return false
	}

	accelerator, err := openstack.NewAcceleratorV1(
		provider,
		openstack.DefaultEndpointOpts,
	)
	if err != nil {
		log.Errorf("failed to create accelerator client: %s", err.Error())
		return false
	}

	devices, err := devices.List(
		accelerator,
		devices.ListOpts{Hostname: definition.Hostname},
	)
	if err != nil {
		log.Errorf("failed to list accelerator devices: %s", err.Error())
		return false
	}

	return len(devices) > 0
}

func GetRoleOverview() (*Role, error) {
	nodes, err := ListNodes()
	if err != nil {
		return nil, err
	}

	role := &Role{}
	for _, node := range nodes {
		switch node.Role {
		case definition.RoleControl:
			role.Control++
		case definition.RoleCompute:
			role.Compute++
		case definition.RoleStorage:
			role.Storage++
		case definition.RoleControlConverged:
			role.ControlConverged++
		default:
			role.Others++
		}
	}

	return role, nil
}

func ListNodes() ([]*definition.Node, error) {
	nodes := []*definition.Node{}

	for _, role := range definition.Roles {
		nodes, err := definition.GetNodesByRole(role)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, nodes...)
	}

	return nodes, nil
}
