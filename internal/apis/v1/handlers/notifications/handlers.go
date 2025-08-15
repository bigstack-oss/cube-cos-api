package notifications

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/notifications/last",
			Func:    GetLastNotification,
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

func GetLastNotification(c *gin.Context) {
	h, err := initHepler(c, "getLastNotification")
	if err != nil {
		log.Errorf("notifications(%s): failed to init helper(%v)", h.reqId, err)
		bodies.SetBadRequest(c, err, nil)
		return
	}

	notification, err := cubecos.GetLastNotification()
	if err != nil {
		log.Errorf("notifications(%s): failed to get last notification: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	msg := "fetch last notification successfully"
	if notification == nil {
		msg = "no last notification found"
	}

	bodies.SetOk(
		c,
		msg,
		notification,
	)
}
