package metrics

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c *gin.Context

	metricGroup  string
	metricType   string
	resourceType string
	reportType   string

	definition.Period
	limit int
	rank
	watch bool
}

type rank struct {
	head int
	tail int
}

func initReqHelper(c *gin.Context) (*helper, error) {
	h := &helper{c: c}
	return h, h.parseParams()
}

func (h *helper) getMetrics() (interface{}, error) {
	switch h.metricGroup {
	case "cpu":
		return h.getCpuMetrics()
	case "memory":
		return h.getMemoryMetrics()
	case "storage":
		return h.getStorageMetrics()
	case "network":
		return h.getNetworkMetrics()
	}

	return nil, fmt.Errorf(
		"invalid metric group(%s) to get metrics",
		h.resourceType,
	)
}
