package services

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version:              api.V1,
			Method:               http.MethodGet,
			Path:                 "/services",
			Func:                 getServices,
			IsNotUnderDataCenter: false,
		},
	}
)

func getServices(c *gin.Context) {
	api.SetStatusOk(
		c,
		"fetch service details successfully",
		cubecos.OrderSensitiveServices,
	)
}
