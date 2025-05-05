package runtime

import (
	"fmt"
	"os"
	"strings"
	"time"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/gin-gonic/gin"
)

func newServiceDiscoveryIdentity() error {
	if v1.DataCenterName == "" {
		return errors.InvalidDataCenterName
	}

	if v1.DataCenterVip == "" {
		return errors.InvalidListenAddress
	}

	v1.ServiceDiscoveryIdentity = fmt.Sprintf(
		"%s-%s-%s",
		v1.DataCenterName,
		v1.DataCenterVip,
		strings.ToLower(v1.DefaultOidcClientSecret[:8]),
	)

	return nil
}

func parseLocalListenAddr() (string, error) {
	if conf.Opts.Spec.Listen.Local == "" {
		conf.Opts.Spec.Listen.Local = v1.ManagementIp
	}

	if conf.Opts.Spec.Listen.Local == "" {
		return "", errors.InvalidListenAddress
	}

	return conf.Opts.Spec.Listen.Local, nil
}

func parseLocalListenPort() (int, error) {
	if conf.Opts.Spec.Listen.Port == 0 {
		return 0, errors.InvalidListenPort
	}

	return conf.Opts.Spec.Listen.Port, nil
}

func parseAdvertisePort() (int, error) {
	if conf.Opts.Spec.Listen.Port == 0 {
		return 0, errors.InvalidListenPort
	}

	return conf.Opts.Spec.Listen.Port, nil
}

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

func newNodeMetadata() error {
	if v1.CurrentRole == "" {
		return errors.InvalidNodeRole
	}

	if v1.Hostname == "" {
		return errors.InvalidHostname
	}

	if v1.DataCenterName == "" {
		return errors.InvalidDataCenterName
	}

	if v1.ManagementIp == "" {
		return errors.InvalidManagementIp
	}

	v1.NodeMetadata = map[string]string{
		"role":         v1.CurrentRole,
		"hostname":     v1.Hostname,
		"dataCenter":   v1.DataCenterName,
		"nodeID":       v1.HostID,
		"serialNumber": v1.SerialNumber,
		"protocol":     conf.Opts.Kind,
		"ip":           v1.ManagementIp,
		"isGpuEnabled": fmt.Sprintf("%t", v1.IsGpuEnabled),
	}
	return nil
}

func genLocalAddr() (string, error) {
	if v1.ListenIp == "" {
		return "", errors.InvalidListenAddress
	}

	if v1.ListenPort == 0 {
		return "", errors.InvalidListenPort
	}

	return fmt.Sprintf(
		"%s:%d",
		v1.ListenIp,
		v1.ListenPort,
	), nil
}

func genServiceDiscoveryAddr() (string, error) {
	if v1.ManagementIp == "" {
		return "", errors.InvalidListenAddress
	}

	if v1.ListenPort == 0 {
		return "", errors.InvalidListenPort
	}

	return fmt.Sprintf(
		"%s:%d",
		v1.ManagementIp,
		v1.ListenPort,
	), nil
}

func genLogoutRedirectUrl() (string, error) {
	if v1.DataCenterVip == "" {
		return "", errors.InvalidListenAddress
	}

	return fmt.Sprintf(
		"https://%s:4443%s",
		v1.DataCenterVip,
		conf.Opts.Spec.Identity.LogoutRedirect,
	), nil
}
