package runtime

import (
	"fmt"
	"os"
	"time"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func parseParams(c *gin.Context) string {
	if c.Request.URL.RawQuery == "" {
		return ""
	}

	return fmt.Sprintf(
		"?%s",
		c.Request.URL.RawQuery,
	)
}

func getLocalTimeZone() string {
	_, offsetSeconds := time.Now().Zone()
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	mins := (offsetSeconds % 3600) / 60
	return fmt.Sprintf(
		"%s%02d:%02d",
		sign,
		hours,
		mins,
	)
}

func getLocalTimeZoneSeconds() int {
	_, offsetSeconds := time.Now().Zone()
	return offsetSeconds
}

func getHostname() (string, error) {
	if conf.Opts.Spec.Identity.Os.Hostname != "" {
		return conf.Opts.Spec.Identity.Os.Hostname, nil
	}

	return os.Hostname()
}

func genNodeMetadata() map[string]string {
	return map[string]string{
		"role":         v1.CurrentRole,
		"hostname":     v1.Hostname,
		"dataCenter":   v1.DataCenterName,
		"nodeID":       v1.HostID,
		"serialNumber": v1.SerialNumber,
		"protocol":     conf.Opts.Kind,
		"ip":           v1.ManagementIp,
		"isGpuEnabled": fmt.Sprintf("%t", v1.IsGpuEnabled),
		"token":        v1.DefaultNodeToken,
	}
}

func genLocalAddr() string {
	return fmt.Sprintf(
		"%s:%d",
		conf.Opts.Spec.Listen.Local,
		conf.Opts.Spec.Listen.Port,
	)
}

func genServiceDiscoveryAddr() string {
	return fmt.Sprintf(
		"%s:%d",
		v1.ManagementIp,
		conf.Opts.Spec.Listen.Port,
	)
}

func genLogoutRedirectUrl() string {
	return fmt.Sprintf(
		"https://%s:4443%s",
		v1.DataCenterVip,
		conf.Opts.Spec.Identity.LogoutRedirect,
	)
}
