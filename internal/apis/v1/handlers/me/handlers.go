package me

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/me",
			Func:    getMe,
		},
	}
)

func getMe(c *gin.Context) {
	username, err := getUsername(c)
	if err != nil {
		log.Errorf("me(%s): %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
	}

	bodies.SetOk(
		c,
		"fetch user info successfully",
		gin.H{"name": username},
	)
}
