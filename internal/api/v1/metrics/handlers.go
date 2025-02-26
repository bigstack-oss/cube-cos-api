package metrics

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
			Path:    "/metrics",
			Func:    getDataCenterSummary,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/metrics/:metricType/:viewType/:entityType",
			Func:    getMetrics,
		},
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/metrics/:metricType/:viewType/:entityType/:entityId",
			Func:    getMetrics,
		},
	}
)

func init() {
	go streamHealth()
}

func getDataCenterSummary(c *gin.Context) {
	h, err := initReqHelper(c, "getDataCenterSummary")
	if err != nil {
		log.Errorf("request(%s): %v", api.GetReqId(c), err)
		api.SetBadRequest(c, err)
		return
	}

	summary, err := h.getDataCenterSummary()
	if err != nil {
		log.Errorf("request(%s): failed to fetch data center summary: %v", api.GetReqId(c), err)
		api.SetInternalServerError(c, err)
		return
	}

	if h.watch {
		watchHealth(h, summary)
		return
	}

	api.SetStatusOk(c, "fetch summary successfully", summary)
}

func getMetrics(c *gin.Context) {
	h, err := initReqHelper(c, "getMetrics")
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

	if h.watch {
		watchHealth(h, metrics)
		return
	}

	api.SetStatusOk(
		c,
		"fetch metrics successfully",
		metrics,
	)
}
