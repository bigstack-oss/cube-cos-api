package ceph

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

const (
	RadosGatewayPort = 8888
)

type SpaceMetrics struct {
	Stats `json:"stats"`
}

type Stats struct {
	TotalBytes      int64 `json:"total_bytes"`
	TotalAvailBytes int64 `json:"total_avail_bytes"`
	TotalUsedBytes  int64 `json:"total_used_bytes"`
}

func GetRadosGatewayUrl() string {
	return fmt.Sprintf("http://%s:%d/", base.DataCenterVip, RadosGatewayPort)
}
