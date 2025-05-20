package queries

import (
	"github.com/gin-gonic/gin"
)

func ParseRecordRequire(c *gin.Context) bool {
	val, found := c.GetQuery("isRecordRequired")
	if !found {
		return true
	}

	return val == "true"
}
