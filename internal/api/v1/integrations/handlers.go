package integrations

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
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
	c.JSON(http.StatusOK, gin.H{
		"code":   http.StatusOK,
		"status": "ok",
		"msg":    "fetch integration list successfully",
		"data":   cubecos.ListBuiltInIntegrations(),
	})
}
