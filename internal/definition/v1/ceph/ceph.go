package ceph

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

const (
	RadosGatewayPort = 8888
)

func GetRadosGatewayUrl() string {
	return fmt.Sprintf("http://%s:%d/", base.DataCenterVip, RadosGatewayPort)
}
