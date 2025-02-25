package metrics

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	handler string

	metricType string
	viewType   string
	entityType string
	entityId   string

	definition.Period
	past string

	limit int
	rank
	watch bool
}

type rank struct {
	head int
	tail int
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	return h, h.parseParams()
}

func (h *helper) getDataCenterSummary() (interface{}, error) {
	return cubecos.GetDataCenterSummary()
}

func (h *helper) getMetrics() (interface{}, error) {
	switch h.metricType {
	case "cpuUsage":
		return h.getCpuUsageMetrics()
	case "memoryUsage":
		return h.getMemoryUsageMetrics()
	case "diskUsage":
		return h.getDiskUsageMetrics()
	case "diskBandwidth":
		return h.getDiskBandwidthMetrics()
	case "diskIops":
		return h.getDiskIopsMetrics()
	case "diskReadIops":
		return h.getDiskReadIopsMetrics()
	case "diskWriteIops":
		return h.getDiskWriteIopsMetrics()
	case "diskLatency":
		return h.getDiskLatencyMetrics()
	case "networkTrafficIn":
		return h.getNetworkTrafficInMetrics()
	case "networkTrafficOut":
		return h.getNetworkTrafficOutMetrics()
	}

	return nil, fmt.Errorf(
		"invalid metric type(%s) to get metrics",
		h.metricType,
	)
}
