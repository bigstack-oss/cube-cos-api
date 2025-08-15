package notifications

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	keyword string
	period  *time.Period
	past    string
	page    *pages.Page
	limit   int
}

func initHepler(c *gin.Context, handler string) (*helper, error) {
	h := helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return &h, h.parseParamByHandler()
}

func (h *helper) listNotifications() (*notificationPage, error) {
	opts, err := h.convertListOpts()
	if err != nil {
		return nil, err
	}

	notifications, err := cubecos.ListNotifications(*opts)
	if err != nil {
		return nil, err
	}

	notifications = h.filterNotifications(notifications)
	return &notificationPage{
		Notifications: h.paginateNotifications(notifications),
		Page:          h.genPageInfo(notifications),
	}, nil
}
