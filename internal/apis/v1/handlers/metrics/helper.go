package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	metricType string
	viewType   string
	entityType string
	entityId   string

	*time.Period
	past            string
	aggregateWindow string

	limit int
	rank
	watch bool
}

type rank struct {
	head int
	tail int
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, reqId: queries.GetReqId(c), handler: handler}
	return h, h.parseParamsByHandler()
}

func (h *helper) getMetrics() (any, error) {
	switch h.metricType {
	case "cpuUsage":
		return h.getCpuUsage()
	case "memoryUsage":
		return h.getMemoryUsage()
	case "diskUsage":
		return h.getDiskUsage()
	case "diskBandwidth":
		return h.getDiskBandwidth()
	case "diskIops":
		return h.getDiskIops()
	case "diskReadIops":
		return h.getDiskReadIops()
	case "diskWriteIops":
		return h.getDiskWriteIops()
	case "diskLatency":
		return h.getDiskLatency()
	case "networkTrafficIn":
		return h.getNetworkIngressTraffic()
	case "networkTrafficOut":
		return h.getNetworkEgressTraffic()
	default:
		return nil, fmt.Errorf(
			"invalid metric type(%s) to get metrics",
			h.metricType,
		)
	}
}
