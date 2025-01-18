package datacenters

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "",
			Func:    getDataCenters,
		},
	}
)

func getDataCenters(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch data center list successfully",
		"data": []definition.DataCenter{
			{
				Name:        definition.Controller,
				VirtualIp:   definition.ControllerVip,
				IsLocal:     true,
				IsHaEnabled: definition.IsHaEnabled,
			},
		},
	})
}
