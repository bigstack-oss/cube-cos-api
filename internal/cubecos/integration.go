package cubecos

import (
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

func ListBuiltInApplications() []integration.Service {
	if !datacenter.IsCloudType() {
		return integration.GetCommonServices()
	}

	return append(
		integration.GetCommonServices(),
		integration.GetCloudService(),
	)
}

func ListBuiltInStorages() []integration.Storage {
	return []integration.Storage{
		{
			Name:         "CubeStorage",
			Type:         "built-in",
			Vendor:       "CubeCOS",
			ManagementIp: base.ManagementIp,
			UpdatedAt:    time.LocalRFC3339(ostime.Now().Local()),
			IsDefault:    true,
			Status: status.Integration{
				Current:      status.Ok,
				IsProcessing: false,
			},
		},
	}
}
