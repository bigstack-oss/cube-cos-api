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
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/events/rank",
			Func:    genEventRank,
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

func getEvents(c *gin.Context) {
	h, err := initHelper(c, "getEvents")
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
	h, err := initHelper(c, "getEventAbstract")
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

func genEventRank(c *gin.Context) {
	h, err := initHelper(c, "genEventRank")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	rank, err := h.genEventRank()
	if err != nil {
		log.Errorf("request(%s): failed to gen event rank: %v", api.GetReqId(c), err)
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
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	conditions, err := h.genEventFilterConditions()
	if err != nil {
		log.Errorf("request(%s): failed to gen event filter conditions: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch event filter conditions successfully",
		conditions,
	)
}
