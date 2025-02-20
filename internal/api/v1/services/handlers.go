package services

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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
	svcs := []v1.Service{}
	for _, svc := range cubecos.OrderSensitiveServices {
		if !svc.IsInternalViewOnly {
			svcs = append(svcs, svc)
		}
	}

	api.SetStatusOk(
		c,
		"fetch service details successfully",
		svcs,
	)
}
