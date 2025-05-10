package opensearch

import (
	"net/http"

	api "github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/opensearch/requests/:requestId",
			Func:    forwardRequestLink,
		},
	}
)

func forwardRequestLink(c *gin.Context) {
	requestId := c.Param("requestId")
	link, err := cubecos.GetOpenSearchRequestLink(requestId)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch request link successfully",
		v1.Dashboard{
			Link:    link,
			Enabled: true,
		},
	)
}
