package metrics

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/metrics"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/metrics",
			Func:    getDataCenterSummary,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/metrics/:metricType/:viewType/:entityType",
			Func:    getMetrics,
		},
		{
			Version: apis.V1,
			Method:  http.MethodGet,
			Path:    "/metrics/:metricType/:viewType/:entityType/:entityId",
			Func:    getMetrics,
		},
	}
)

func init() {
	go streamingWatcher()
}

func getDataCenterSummary(c *gin.Context) {
	h, err := initHelper(c, "getDataCenterSummary")
	if err != nil {
		log.Errorf("metrics(%s): %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	summary := cubecos.GetMetricsSummary()
	if h.watch {
		watchHealth(h, summary)
		return
	}

	bodies.SetOk(
		c,
		"fetch summary successfully",
		summary,
	)
}

func getMetrics(c *gin.Context) {
	h, err := initHelper(c, "getMetrics")
	if err != nil {
		log.Errorf("metrics(%s): failed to init helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	metrics, err := h.getMetrics()
	if err != nil {
		log.Errorf("metrics(%s): failed to fetch %s: %v", h.reqId, h.metricType, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if h.watch {
		watchHealth(h, metrics)
		return
	}

	bodies.SetOk(
		c,
		"fetch metrics successfully",
		metrics,
	)
}
