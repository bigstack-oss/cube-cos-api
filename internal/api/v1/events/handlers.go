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
	err := checkQueryParams(c)
	if err != nil {
		log.Errorf("request(%s): invalid query params: %v", api.GetReqId(c), err)
		api.SetErrBadRequestResp(c, err)
		return
	}

	events, err := cubecos.ListEvents(genQueryStmt(c))
	if err != nil {
		log.Errorf("request(%s): failed to fetch events: %v", api.GetReqId(c), err)
		api.SetErrInternalServerErrorResp(c, err)
		return
	}

	api.SetStatusOkResp(
		c,
		"fetch events successfully",
		events,
	)
}
