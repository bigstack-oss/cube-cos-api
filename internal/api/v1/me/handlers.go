package me

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/me",
			Func:    getMe,
		},
	}
)

func getMe(c *gin.Context) {
	username := samlsp.AttributeFromContext(c.Request.Context(), "username")
	api.SetStatusOk(
		c,
		"fetch own user info successfully",
		gin.H{"name": username},
	)
}
