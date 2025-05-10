package queries

import "github.com/gin-gonic/gin"

func ParseClusterWise(c *gin.Context) bool {
	queries := c.Request.URL.Query()
	if len(queries) == 0 {
		return true
	}

	_, found := queries["clusterWise"]
	if !found {
		return true
	}

	return c.DefaultQuery("clusterWise", "false") == "true"
}
