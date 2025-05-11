package grafana

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/grafana"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/hosts/:hostname",
			Func:    forwardHostLink,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/instances/:instanceId",
			Func:    forwardInstanceLink,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/topHosts",
			Func:    forwardTopHostLink,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/topInstances",
			Func:    forwardTopInstanceLink,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/networks",
			Func:    forwardNetworksLink,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/networkDevices",
			Func:    forwardNetworkDevicesLink,
		},
		{
			Version: apis.V1,
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
		grafana.Dashboard{
			Link:    genHostLink(c),
			Enabled: true,
		},
	)
}

func forwardInstanceLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch instance link successfully",
		grafana.Dashboard{
			Link:    genInstanceLink(c),
			Enabled: true,
		},
	)
}

func forwardTopHostLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch top host link successfully",
		grafana.Dashboard{
			Link:    genTopHostLink(),
			Enabled: true,
		},
	)
}

func forwardTopInstanceLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch top instance link successfully",
		grafana.Dashboard{
			Link:    genTopInstanceLink(),
			Enabled: true,
		},
	)
}

func forwardNetworksLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch networks link successfully",
		grafana.Dashboard{
			Link:    genNetworksLink(),
			Enabled: cubecos.IsOvnSFlowEnabled(),
		},
	)
}

func forwardNetworkDevicesLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch network devices link successfully",
		grafana.Dashboard{
			Link:    genNetworkDevicesLink(),
			Enabled: cubecos.IsOvnSFlowEnabled(),
		},
	)
}

func forwardStoragesLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch storages link successfully",
		grafana.Dashboard{
			Link:    genStoragesLink(),
			Enabled: true,
		},
	)
}
