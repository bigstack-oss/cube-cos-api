package grafana

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
			Path:    "/grafana/networkDevices",
			Func:    forwardNetworkDevicesLink,
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
	bodies.SetOk(
		c,
		"fetch host link successfully",
		v1.Dashboard{
			Link:    genHostLink(c),
			Enabled: true,
		},
	)
}

func forwardInstanceLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch instance link successfully",
		v1.Dashboard{
			Link:    genInstanceLink(c),
			Enabled: true,
		},
	)
}

func forwardTopHostLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch top host link successfully",
		v1.Dashboard{
			Link:    genTopHostLink(),
			Enabled: true,
		},
	)
}

func forwardTopInstanceLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch top instance link successfully",
		v1.Dashboard{
			Link:    genTopInstanceLink(),
			Enabled: true,
		},
	)
}

func forwardNetworksLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch networks link successfully",
		v1.Dashboard{
			Link:    genNetworksLink(),
			Enabled: cubecos.IsOvnSFlowEnabled(),
		},
	)
}

func forwardNetworkDevicesLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch network devices link successfully",
		v1.Dashboard{
			Link:    genNetworkDevicesLink(),
			Enabled: cubecos.IsOvnSFlowEnabled(),
		},
	)
}

func forwardStoragesLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch storages link successfully",
		v1.Dashboard{
			Link:    genStoragesLink(),
			Enabled: true,
		},
	)
}
