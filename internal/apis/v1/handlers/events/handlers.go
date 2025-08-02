package events

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Module   = "events"
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/events",
			Func:    listEvents,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/events/predefined",
			Func:    listPredefinedEvents,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/events/abstract",
			Func:    listEventAbstract,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/events/rank",
			Func:    getEventRank,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/events/filterConditions",
			Func:    getEventFilterConditions,
		},
	}
)

func init() {
	go streamWatchers()
}

func listEvents(c *gin.Context) {
	h, err := initHelper(c, "listEvents")
	if err != nil {
		log.Errorf("events(%s): %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	events, err := h.listEvents()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, events)
		return
	}

	bodies.SetOk(
		c,
		"fetch events successfully",
		events,
	)
}

func listPredefinedEvents(c *gin.Context) {
	h, err := initHelper(c, "listPredefinedEvents")
	if err != nil {
		log.Errorf("events(%s): %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	events, err := h.listPredefinedEvents()
	if err != nil {
		log.Errorf("events(%s): failed to list predefined events: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch predefined events successfully",
		events,
	)
}

func listEventAbstract(c *gin.Context) {
	h, err := initHelper(c, "listEventAbstract")
	if err != nil {
		log.Errorf("events(%s): %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	abstract, err := h.listEventAbstract()
	if err != nil {
		log.Errorf("events(%s): failed to gen event abstract: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, abstract)
		return
	}

	bodies.SetOk(
		c,
		"fetch event abstract successfully",
		abstract,
	)
}

func getEventRank(c *gin.Context) {
	h, err := initHelper(c, "getEventRank")
	if err != nil {
		log.Errorf("events(%s): %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	rank, err := h.getEventRank()
	if err != nil {
		log.Errorf("events(%s): failed to gen event rank: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchEvents(h, rank)
		return
	}

	bodies.SetOk(
		c,
		"fetch event rank successfully",
		rank,
	)
}

func getEventFilterConditions(c *gin.Context) {
	h, err := initHelper(c, "getEventFilterConditions")
	if err != nil {
		log.Errorf("events(%s): %v", queries.GetReqId(c), err)
		bodies.SetBadRequest(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch event filter conditions successfully",
		h.getEventFilterConditions(),
	)
}
