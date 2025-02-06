package events

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
			Path:    "/events",
			Func:    getEvents,
		},
	}
)

func init() {
	go streamEvents()
}

func getEvents(c *gin.Context) {
	h, err := initReqHelper(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	resp, err := h.genEventResp()
	if err != nil {
		log.Errorf("request(%s): failed to gen event resp: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, resp)
		return
	}

	api.SetStatusOk(
		c,
		"fetch events successfully",
		resp,
	)
}
