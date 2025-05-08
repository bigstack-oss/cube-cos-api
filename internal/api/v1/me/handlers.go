package me

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
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
	username, err := getUsername(c)
	if err != nil {
		log.Errorf("me(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
	}

	api.SetStatusOk(
		c,
		"fetch user info successfully",
		gin.H{"name": username},
	)
}
