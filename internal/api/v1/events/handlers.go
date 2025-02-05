package events

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events",
			Func:    getEvents,
		},
	}
)

func getEvents(c *gin.Context) {
	h, err := initReqHelper(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	stmt := h.genQueryStmt()
	events, err := cubecos.GetEvents(stmt)
	if err != nil {
		log.Errorf("request(%s): failed to fetch events: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	page, err := h.genPageInfo(events)
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch events successfully",
		data{
			Events: events,
			Page:   page,
		},
	)
}
