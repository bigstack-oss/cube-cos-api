package cubecos

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1/accelerators/devices"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	log "go-micro.dev/v5/logger"
)

func IsGpuEnabled() (bool, error) {
	provider, err := openstack.NewProvider(conf.Opts.Spec.ResourceControl.Openstack.Auth.File)
	if err != nil {
		log.Errorf("cos: failed to create openstack provider: %s", err.Error())
		return false, err
	}

	accelerator, err := openstack.NewAcceleratorV1(
		provider,
		openstack.DefaultEndpointOpts,
	)
	if err != nil {
		log.Errorf("cos: failed to create accelerator client: %s", err.Error())
		return false, err
	}

	devices, err := devices.List(
		accelerator,
		devices.ListOpts{Hostname: base.Hostname},
	)
	if err != nil {
		log.Errorf("cos: failed to list accelerator devices: %s", err.Error())
		return false, err
	}

	return len(devices) > 0, nil
}
