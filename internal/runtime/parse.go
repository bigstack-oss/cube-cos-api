package runtime

import (
	"fmt"
	"os"
	"strings"
	"time"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/gin-gonic/gin"
)

func newServiceDiscoveryIdentity() error {
	if v1.DataCenterName == "" {
		return errors.ErrInvalidDataCenterName
	}

	if v1.DataCenterVip == "" {
		return errors.ErrInvalidListenAddress
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
		return "", errors.ErrInvalidListenAddress
	}

	return conf.Opts.Spec.Listen.Local, nil
}

func parseLocalListenPort() (int, error) {
	if conf.Opts.Spec.Listen.Port == 0 {
		return 0, errors.ErrInvalidListenPort
	}

	return conf.Opts.Spec.Listen.Port, nil
}

func parseAdvertisePort() (int, error) {
	if conf.Opts.Spec.Listen.Port == 0 {
		return 0, errors.ErrInvalidListenPort
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
		return errors.ErrInvalidNodeRole
	}

	if v1.Hostname == "" {
		return errors.ErrInvalidHostname
	}

	if v1.DataCenterName == "" {
		return errors.ErrInvalidDataCenterName
	}

	if v1.ManagementIp == "" {
		return errors.ErrInvalidManagementIp
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
		return "", errors.ErrInvalidListenAddress
	}

	if v1.ListenPort == 0 {
		return "", errors.ErrInvalidListenPort
	}

	return fmt.Sprintf(
		"%s:%d",
		v1.ListenIp,
		v1.ListenPort,
	), nil
}

func genServiceDiscoveryAddr() (string, error) {
	if v1.ManagementIp == "" {
		return "", errors.ErrInvalidListenAddress
	}

	if v1.ListenPort == 0 {
		return "", errors.ErrInvalidListenPort
	}

	return fmt.Sprintf(
		"%s:%d",
		v1.ManagementIp,
		v1.ListenPort,
	), nil
}

func genLogoutRedirectUrl() (string, error) {
	if v1.DataCenterVip == "" {
		return "", errors.ErrInvalidListenAddress
	}

	return fmt.Sprintf(
		"https://%s:4443%s",
		v1.DataCenterVip,
		conf.Opts.Spec.Identity.Redirect,
	), nil
}

func genRequestMsg(c *gin.Context) string {
	return fmt.Sprintf("%s %s%s", c.Request.Method, c.Request.URL.Path, parseParams(c))
}

func parseRedirectPath() (string, error) {
	if conf.Opts.Spec.Identity.Redirect != "" {
		return conf.Opts.Spec.Identity.Redirect, nil
	}

	if v1.DefaultRedirectPath != "" {
		return v1.DefaultRedirectPath, nil
	}

	return "", errors.ErrNoRedirectPathFound
}
