package datacenters

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []apis.Handler{
		{
			Version:              apis.V1,
			Method:               http.MethodGet,
			Path:                 "/datacenters",
			Func:                 listDataCenters,
			IsNotUnderDataCenter: true,
		},
		{
			Version:              apis.V1,
			Method:               http.MethodGet,
			Path:                 "/datacenters/:DataCenter",
			Func:                 getDataCenter,
			IsNotUnderDataCenter: true,
		},
		{
			Version:              apis.V1,
			Method:               http.MethodPost,
			Path:                 "/datacenters/:DataCenter/rollingReboot",
			Func:                 softRebootDataCenter,
			IsNotUnderDataCenter: true,
		},
	}
)

// M2 plan: the data center info will be persisted and retrieved from the database
// M1 only has one data center, so just return the current one
func listDataCenters(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch data center list successfully",
		[]base.DataCenter{
			getLocalDataCenter(),
		},
	)
}

func getDataCenter(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch data center list successfully",
		getLocalDataCenter(),
	)
}

func softRebootDataCenter(c *gin.Context) {
	err := cubecos.PowerCycleDataCenter()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetAccepted(
		c,
		"rollout data center by soft reboot successfully",
	)
}
