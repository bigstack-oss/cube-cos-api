package runtime

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/gin-gonic/gin"
)

func isLiveCheck(c *gin.Context) bool {
	return c.Request.Method == "GET" && c.Request.URL.Path == "/live"
}

func filterReq(c *gin.Context) {
	if isGetSamlAcs(c) {
		c.AbortWithStatus(403)
		return
	}

	c.Next()
}

type route struct {
	method string
	path   string
}

// authFreeRoutes lists every route served without credentials, keyed by the
// route template gin matched rather than the raw request path.
var authFreeRoutes = map[route]bool{
	{http.MethodGet, apis.V1 + "/datacenters"}:                             true,
	{http.MethodGet, dcRoute("/grafana/hosts/:hostname")}:                  true,
	{http.MethodGet, dcRoute("/grafana/instances/:instanceId")}:            true,
	{http.MethodGet, dcRoute("/grafana/topHosts")}:                         true,
	{http.MethodGet, dcRoute("/grafana/topInstances")}:                     true,
	{http.MethodGet, dcRoute("/grafana/networks")}:                         true,
	{http.MethodGet, dcRoute("/grafana/networkDevices")}:                   true,
	{http.MethodGet, dcRoute("/grafana/storages")}:                         true,
	{http.MethodGet, dcRoute("/grafana/devices/:hostname/gpuUtilization")}: true,
	{http.MethodGet, dcRoute("/grafana/devices/:hostname/gpuVram")}:        true,
	{http.MethodGet, dcRoute("/opensearch/requests/:requestId")}:           true,
}

func dcRoute(path string) string {
	return getUrlParentPath(apis.Handler{Version: apis.V1}) + path
}

func isAuthFreeReq(c *gin.Context) bool {
	return authFreeRoutes[route{c.Request.Method, c.FullPath()}]
}

func isGetSamlAcs(c *gin.Context) bool {
	return c.Request.Method == "GET" &&
		c.Request.URL.Path == "/saml/acs"
}
