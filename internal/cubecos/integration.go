package cubecos

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
)

func ListBuiltInIntegrations() []integration.Service {
	if !datacenter.IsCloudType() {
		return integration.GetCommonServices()
	}

	return append(
		integration.GetCommonServices(),
		integration.GetCloudService(),
	)
}
