package queries

import "github.com/gin-gonic/gin"

func GetReqId(c *gin.Context) string {
	id, found := c.Get("reqId")
	if !found {
		return ""
	}

	return id.(string)
}
