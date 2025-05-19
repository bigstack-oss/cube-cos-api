package runtime

import (
	"strings"

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

func isAuthFreeReq(c *gin.Context) bool {
	if isGetDataCenters(c) {
		return true
	}

	if isGetGrafana(c) {
		return true
	}

	if isGetOpenSearch(c) {
		return true
	}

	return false
}

func isGetSamlAcs(c *gin.Context) bool {
	return c.Request.Method == "GET" &&
		c.Request.URL.Path == "/saml/acs"
}

func isGetDataCenters(c *gin.Context) bool {
	return c.Request.Method == "GET" &&
		c.Request.URL.Path == "/api/v1/datacenters"
}

func isGetGrafana(c *gin.Context) bool {
	return c.Request.Method == "GET" &&
		strings.Contains(c.Request.URL.Path, "/grafana")
}

func isGetOpenSearch(c *gin.Context) bool {
	return c.Request.Method == "GET" &&
		strings.Contains(c.Request.URL.Path, "/opensearch")
}
