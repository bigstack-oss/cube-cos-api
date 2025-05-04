package opensearch

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
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch request link successfully",
		v1.Dashboard{
			Link:    link,
			Enabled: true,
		},
	)
}
