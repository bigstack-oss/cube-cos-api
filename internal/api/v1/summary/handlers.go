package summary

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
			Path:    "/summary",
			Func:    getSummary,
		},
	}
)

func init() {
	go onDemandStreamSummary()
}

func getSummary(c *gin.Context) {
	summary, err := cubecos.GetDataCenterSummary()
	if err != nil {
		log.Errorf("request(%s): failed to fetch data center summary: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	watch, err := parseWatch(c)
	if err != nil {
		log.Errorf("request(%s): failed to parse watch query: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	if !watch {
		api.SetStatusOk(c, "fetch summary successfully", summary)
		return
	}

	watchSummary(c, summary)
}
