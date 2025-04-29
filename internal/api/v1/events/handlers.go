package events

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Module   = "events"
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events",
			Func:    listEvents,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events/abstract",
			Func:    listEventAbstract,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events/rank",
			Func:    getEventRank,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events/filterConditions",
			Func:    getEventFilterConditions,
		},
	}
)

func init() {
	go streamingWatcher()
}

func listEvents(c *gin.Context) {
	h, err := initHelper(c, "listEvents")
	if err != nil {
		log.Errorf("events(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	events, err := h.listEvents()
	if err != nil {
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

func listEventAbstract(c *gin.Context) {
	h, err := initHelper(c, "listEventAbstract")
	if err != nil {
		log.Errorf("events(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	abstract, err := h.listEventAbstract()
	if err != nil {
		log.Errorf("events(%s): failed to gen event abstract: %v", api.GetReqId(c), err)
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

func getEventRank(c *gin.Context) {
	h, err := initHelper(c, "getEventRank")
	if err != nil {
		log.Errorf("events(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	rank, err := h.getEventRank()
	if err != nil {
		log.Errorf("events(%s): failed to gen event rank: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, rank)
		return
	}

	api.SetStatusOk(
		c,
		"fetch event rank successfully",
		rank,
	)
}

func getEventFilterConditions(c *gin.Context) {
	h, err := initHelper(c, "getEventFilterConditions")
	if err != nil {
		log.Errorf("events(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch event filter conditions successfully",
		h.getEventFilterConditions(),
	)
}
