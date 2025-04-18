package cubecos

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1/accelerators/devices"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
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

func IsGpuEnabled() bool {
	provider, err := openstack.NewProvider(conf.Opts.Spec.ResourceControl.Openstack.Auth.File)
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
		devices.ListOpts{Hostname: v1.Hostname},
	)
	if err != nil {
		log.Errorf("failed to list accelerator devices: %s", err.Error())
		return false
	}

	return len(devices) > 0
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
