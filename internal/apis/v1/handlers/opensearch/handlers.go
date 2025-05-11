package opensearch

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/opensearch"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/opensearch/requests/:requestId",
			Func:    forwardRequestLink,
		},
	}
)

func forwardRequestLink(c *gin.Context) {
	id := c.Param("requestId")
	link, err := cubecos.GetOpenSearchRequestLink(id)
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch request link successfully",
		opensearch.Dashboard{
			Link:    link,
			Enabled: true,
		},
	)
}
