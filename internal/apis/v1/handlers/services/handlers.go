package services

import (
	"net/http"

	api "github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/services"
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
	svcs := []services.Service{}
	for _, svc := range cubecos.OrderSensitiveServices {
		if !svc.IsInternalViewOnly {
			svcs = append(svcs, svc)
		}
	}

	bodies.SetOk(
		c,
		"fetch service details successfully",
		svcs,
	)
}
