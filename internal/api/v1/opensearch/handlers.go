package opensearch

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/opensearch/instances/:instanceId",
			Func:    forwardInstanceLink,
		},
	}
)

func forwardInstanceLink(c *gin.Context) {
	api.SetStatusOk(
		c,
		"fetch instance link successfully",
		v1.Dashboard{
			Link:    genInstanceLink(c),
			Enabled: true,
		},
	)
}
