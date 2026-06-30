package grafana

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/grafana"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
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
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/devices/:hostname/gpuUtilization",
			Func:    forwardDeviceGpuUtilizationLink,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/grafana/devices/:hostname/gpuVram",
			Func:    forwardDeviceGpuVramLink,
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

// Returns the device dashboard deep-link for a physical node's GPU utilization
// history (panel 50), filtered by var-GPU_HOST, whose value must equal the
// gpu.host `host` tag (verified equal to Node.Hostname). Enabled is gated on
// node existence (cluster-wide); GetNodeGpusMap is NOT used here because it is
// local-only and would report the wrong node for remote ones.
func forwardDeviceGpuUtilizationLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch device gpu utilization link successfully",
		grafana.Dashboard{
			Link:    genGpuUtilizationHistoryLink(c),
			Enabled: nodes.IsExist(c.Param("hostname")),
		},
	)
}

// Returns the device dashboard deep-link for a physical node's GPU VRAM usage
// history (panel 51). Same var-GPU_HOST / enabled semantics as the utilization link.
func forwardDeviceGpuVramLink(c *gin.Context) {
	bodies.SetOk(
		c,
		"fetch device gpu vram link successfully",
		grafana.Dashboard{
			Link:    genGpuVramHistoryLink(c),
			Enabled: nodes.IsExist(c.Param("hostname")),
		},
	)
}
