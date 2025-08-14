package notifications

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/notifications",
			Func:    ListNotifications,
		},
	}
)

func ListNotifications(c *gin.Context) {
	h, err := initHepler(c, "listNotifications")
	if err != nil {
		log.Errorf("notifications(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	notifications, err := h.listNotifications()
	if err != nil {
		log.Errorf("notifications(%s): failed to list notifications: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch notifications successfully",
		notifications,
	)
}
