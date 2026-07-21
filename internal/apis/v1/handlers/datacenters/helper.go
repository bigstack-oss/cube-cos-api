package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	log "go-micro.dev/v5/logger"
)

// Resolved per request: the cached base.ActiveFirmwareVersion is this node's
// version as of service start, which lies about a cluster mid-upgrade.
func getClusterFirmwareVersion() string {
	version, err := cubecos.GetClusterFirmwareVersion()
	if err != nil {
		log.Errorf("datacenters: failed to get cluster firmware version(%v)", err)
		return base.ActiveFirmwareVersion
	}

	return version
}

func getLocalDataCenter() base.DataCenter {
	version := getClusterFirmwareVersion()
	return base.DataCenter{
		Type:        datacenter.GetType(),
		Roles:       datacenter.GetAllowRoles(),
		Name:        base.DataCenterName,
		Version:     version,
		VirtualIp:   base.DataCenterVip,
		IsLocal:     true,
		IsHaEnabled: base.IsHaEnabled,
		UtcTimeZone: time.LocalZone,
		Firmware: base.Firmware{
			Version:   version,
			UpdatedAt: base.ActiveFirmwareUpdatedAt,
		},
		Fixpack: base.Fixpack{
			Name:      base.FixpackName,
			Version:   base.FixpackVersion,
			UpdatedAt: base.FixpackUpdatedAt,
		},
		Additional: base.Additional{
			HelpUrl:           base.DataCenterHelpUrl,
			V1ApiDocUrl:       base.GenApiDocUrl(),
			NodeLicenseStatus: getNodeLicenseStatus(),
		},
	}
}
