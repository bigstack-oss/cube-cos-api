package grafana

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/gin-gonic/gin"
)

func genHostLink(c *gin.Context) string {
	return fmt.Sprintf(
		"https://%s/grafana/d/i-R2q81iz/host?refresh=5m&kiosk=tv&orgId=1&var-HOST=%s",
		base.DataCenterVip,
		c.Param("hostname"),
	)
}

func genInstanceLink(c *gin.Context) string {
	return fmt.Sprintf(
		"https://%s/grafana/d/PVW6vU7Wz/instance?refresh=5m&kiosk=tv&orgId=1&var-UUID=%s",
		base.DataCenterVip,
		c.Param("instanceId"),
	)
}

func genTopHostLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/M3ncw6lmk/top-hosts?refresh=5m&kiosk=tv&orgId=1",
		base.DataCenterVip,
	)
}

func genTopInstanceLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/qzfq087Wk/top-instances?refresh=5m&orgId=1&var-TID=&var-TOP=50&var-TENANT=admin",
		base.DataCenterVip,
	)
}

func genNetworksLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/Xx2kkftWk/network?orgId=1&refresh=5m",
		base.DataCenterVip,
	)
}

func genNetworkDevicesLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/i-device/device?refresh=5m&orgId=1",
		base.DataCenterVip,
	)
}

func genStoragesLink() string {
	return fmt.Sprintf(
		"https://%s/grafana/d/QTc_sAxiw/storage?refresh=5m&kiosk=tv&orgId=1",
		base.DataCenterVip,
	)
}
