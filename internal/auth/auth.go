package auth

import (
	"github.com/gin-gonic/gin"
)

func VerifyReq() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Verify request
		c.Next()
	}
}
