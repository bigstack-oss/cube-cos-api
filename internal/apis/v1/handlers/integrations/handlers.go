package integrations

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations",
			Func:    listApplications,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/applications",
			Func:    listApplications,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/integrations/storages",
			Func:    listStorages,
		},
	}
)

func listApplications(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch integrated applications successfully",
		cubecos.ListBuiltInApplications(),
	)
}

func listStorages(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch integrated storages successfully",
		cubecos.ListBuiltInStorages(),
	)
}
