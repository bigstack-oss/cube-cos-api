package metrics

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
			Path:    "/metrics",
			Func:    getDataCenterMetricsSummary,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/metrics/:metricGroup/:resourceType",
			Func:    getMetrics,
		},
	}
)

func init() {
	go streamSummary()
}

func getDataCenterMetricsSummary(c *gin.Context) {
	watch, err := parseWatch(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	summary, err := cubecos.GetDataCenterSummary()
	if err != nil {
		log.Errorf("request(%s): failed to fetch data center summary: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if watch {
		watchSummary(c, summary)
		return
	}

	api.SetStatusOk(c, "fetch summary successfully", summary)
}

func getMetrics(c *gin.Context) {
	h, err := initReqHelper(c)
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	metrics, err := h.getMetrics()
	if err != nil {
		log.Errorf("request(%s): failed to fetch metrics: %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch metrics successfully",
		metrics,
	)
}
