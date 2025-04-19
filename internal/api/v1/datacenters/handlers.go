package datacenters

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version:              api.V1,
			Method:               http.MethodGet,
			Path:                 "/datacenters",
			Func:                 getDataCenters,
			IsNotUnderDataCenter: true,
		},
		{
			Version:              api.V1,
			Method:               http.MethodGet,
			Path:                 "/datacenters/:DataCenter",
			Func:                 getDataCenter,
			IsNotUnderDataCenter: true,
		},
	}
)

// TODO M2: the data center info will be persisted and retrieved from the database
// M1 only has one data center, so just return the current one
func getDataCenters(c *gin.Context) {
	api.SetStatusOk(
		c,
		"fetch data center list successfully",
		[]v1.DataCenter{
			{
				Name:        v1.DataCenterName,
				Version:     v1.DataCenterVersion,
				VirtualIp:   v1.DataCenterVip,
				IsLocal:     true,
				IsHaEnabled: v1.IsHaEnabled,
				UtcTimeZone: v1.LocalTimeZone,
				Additional: v1.Additional{
					HelpUrl: v1.DataCenterHelpUrl,
				},
			},
		},
	)
}

func getDataCenter(c *gin.Context) {
	api.SetStatusOk(
		c,
		"fetch data center list successfully",
		v1.DataCenter{
			Name:        v1.DataCenterName,
			Version:     v1.DataCenterVersion,
			VirtualIp:   v1.DataCenterVip,
			IsLocal:     true,
			IsHaEnabled: v1.IsHaEnabled,
			UtcTimeZone: v1.LocalTimeZone,
			Additional: v1.Additional{
				HelpUrl:       v1.DataCenterHelpUrl,
				LicenseStatus: getLicenseStatus(),
			},
		},
	)
}
