package notifications

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	period *time.Period
	past   string
	limit  int
}

func initHepler(c *gin.Context, handler string) (*helper, error) {
	h := helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return &h, h.parseListNotificationParams()
}

func (h *helper) listNotifications() ([]notifications.Notification, error) {
	opts, err := h.convertListOpts()
	if err != nil {
		return nil, err
	}

	return cubecos.ListNotifications(*opts)
}
