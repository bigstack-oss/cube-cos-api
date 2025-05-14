package cubecos

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
)

func ListBuiltInIntegrations() []integration.Service {
	if !datacenter.IsCloudType() {
		return integration.Common
	}

	return append(
		integration.Common,
		integration.Cloud,
	)
}
