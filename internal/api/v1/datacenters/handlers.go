package datacenters

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/datacenters",
			Func:    getDataCenters,
		},
	}
)

func getDataCenters(c *gin.Context) {
	controller, err := cubecos.ReadHexTuning(cubecos.CubeSysController)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":   http.StatusInternalServerError,
			"status": "internal server error",
			"msg":    "failed to read data center info",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch data center list successfully",
		"data": []definition.DataCenter{
			{
				Name:        controller,
				VirtualIp:   definition.ControllerVip,
				IsLocal:     true,
				IsHaEnabled: definition.IsHaEnabled,
			},
		},
	})
}
