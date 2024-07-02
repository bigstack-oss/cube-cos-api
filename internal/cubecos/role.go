package cubecos

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/accelerators/devices"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

const (
	cubeSysRole = "cubesys.role"
)

func GetNodeRole() (string, error) {
	role, err := HexTuningRead(cubeSysRole)
	if err != nil {
		return "", err
	}

	if role == "" {
		return "", fmt.Errorf("role is empty")
	}

	return role, nil
}

func IsGPUEnabled() bool {
	provider, err := openstack.NewProvider(config.Data.Spec.Dependency.Openstack)
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
