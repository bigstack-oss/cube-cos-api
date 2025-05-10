package integrations

import (
	"net/http"

	api "github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/integrations",
			Func:    getIntegrations,
		},
	}
)

func getIntegrations(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch integrations successfully",
		cubecos.ListBuiltInIntegrations(),
	)
}
