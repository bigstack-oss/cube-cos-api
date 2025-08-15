package datacenters

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/datacenter"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

func getLocalDataCenter() base.DataCenter {
	return base.DataCenter{
		Type:        datacenter.GetType(),
		Roles:       datacenter.GetAllowRoles(),
		Name:        base.DataCenterName,
		Version:     base.ActiveFirmwareVersion,
		VirtualIp:   base.DataCenterVip,
		IsLocal:     true,
		IsHaEnabled: base.IsHaEnabled,
		UtcTimeZone: time.LocalZone,
		Firmware: base.System{
			Version:   base.ActiveFirmwareVersion,
			UpdatedAt: base.ActiveFirmwareUpdatedAt,
		},
		Fixpack: base.System{
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
