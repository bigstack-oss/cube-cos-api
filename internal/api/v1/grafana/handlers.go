package grafana

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/hosts/:hostname",
			Func:    forwardHostLink,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/instances/:instanceId",
			Func:    forwardInstanceLink,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/topHosts",
			Func:    forwardTopHostLink,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/topInstances",
			Func:    forwardTopInstanceLink,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/networks",
			Func:    forwardNetworksLink,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/storages",
			Func:    forwardStoragesLink,
		},
	}
)

func forwardHostLink(c *gin.Context) {
	c.Redirect(http.StatusFound, genHostLink(c))
}

func forwardInstanceLink(c *gin.Context) {
	c.Redirect(http.StatusFound, genInstanceLink(c))
}

func forwardTopHostLink(c *gin.Context) {
	c.Redirect(http.StatusFound, genTopHostLink())
}

func forwardTopInstanceLink(c *gin.Context) {
	c.Redirect(http.StatusFound, genTopInstanceLink())
}

func forwardNetworksLink(c *gin.Context) {
	c.Redirect(http.StatusFound, genNetworksLink())
}

func forwardStoragesLink(c *gin.Context) {
	c.Redirect(http.StatusFound, genStoragesLink())
}
