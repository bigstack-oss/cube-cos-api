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
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events/abstract",
			Func:    getEventAbstract,
		},
	}
)

func init() {
	go streamEvents()
}

func getEvents(c *gin.Context) {
	h, err := initReqHelper(c, "getEvents")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	events, err := h.genEvents()
	if err != nil {
		log.Errorf("request(%s): failed to gen events: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, events)
		return
	}

	api.SetStatusOk(
		c,
		"fetch events successfully",
		events,
	)
}

func getEventAbstract(c *gin.Context) {
	h, err := initReqHelper(c, "getEventAbstract")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	abstract, err := h.genEventAbstract()
	if err != nil {
		log.Errorf("request(%s): failed to gen event abstract: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, abstract)
		return
	}

	api.SetStatusOk(
		c,
		"fetch event abstract successfully",
		abstract,
	)
}
