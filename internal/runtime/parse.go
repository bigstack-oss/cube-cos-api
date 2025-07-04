package runtime

import (
	"fmt"
	"os"
	"strings"
	"time"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"

	"github.com/gin-gonic/gin"
)

func newServiceDiscoveryIdentity() error {
	if base.DataCenterName == "" {
		return errors.ErrInvalidDataCenterName
	}

	if base.DataCenterVip == "" {
		return errors.ErrInvalidListenAddress
	}

	base.ServiceDiscoveryIdentity = fmt.Sprintf(
		"%s-%s-%s",
		base.DataCenterName,
		base.DataCenterVip,
		strings.ToLower(auths.DefaultOidcClientSecret[:8]),
	)

	return nil
}

func parseLocalListenAddr() (string, error) {
	if conf.Opts.Spec.Listen.Local == "" {
		conf.Opts.Spec.Listen.Local = base.ManagementIp
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
	if base.CurrentRole == "" {
		return errors.ErrInvalidNodeRole
	}

	if base.Hostname == "" {
		return errors.ErrInvalidHostname
	}

	if base.DataCenterName == "" {
		return errors.ErrInvalidDataCenterName
	}

	if base.ManagementIp == "" {
		return errors.ErrInvalidManagementIp
	}

	base.NodeMetadata = map[string]string{
		"role":         base.CurrentRole,
		"hostname":     base.Hostname,
		"dataCenter":   base.DataCenterName,
		"nodeID":       base.HostID,
		"serialNumber": base.SerialNumber,
		"boardSerial":  base.BoardSerial,
		"protocol":     conf.Opts.Kind,
		"ip":           base.ManagementIp,
		"isGpuEnabled": fmt.Sprintf("%t", base.IsGpuEnabled),
	}

	return nil
}

func genLocalAddr() (string, error) {
	if base.ListenIp == "" {
		return "", errors.ErrInvalidListenAddress
	}

	if base.ListenPort == 0 {
		return "", errors.ErrInvalidListenPort
	}

	return fmt.Sprintf(
		"%s:%d",
		base.ListenIp,
		base.ListenPort,
	), nil
}

func genServiceDiscoveryAddr() (string, error) {
	if base.ManagementIp == "" {
		return "", errors.ErrInvalidListenAddress
	}

	if base.ListenPort == 0 {
		return "", errors.ErrInvalidListenPort
	}

	return fmt.Sprintf(
		"%s:%d",
		base.ManagementIp,
		base.ListenPort,
	), nil
}

func genLogoutRedirectUrl() (string, error) {
	if base.DataCenterVip == "" {
		return "", errors.ErrInvalidListenAddress
	}

	return fmt.Sprintf(
		"https://%s:%d%s",
		base.DataCenterVip,
		conf.Opts.Spec.Identity.Saml.ServiceProvider.Host.Port,
		conf.Opts.Spec.Identity.Redirect,
	), nil
}

func genReqMsg(c *gin.Context) string {
	return fmt.Sprintf(
		"%s %s%s",
		c.Request.Method,
		c.Request.URL.Path,
		parseParams(c),
	)
}

func parseRedirectPath() (string, error) {
	if conf.Opts.Spec.Identity.Redirect != "" {
		return conf.Opts.Spec.Identity.Redirect, nil
	}

	if auths.DefaultRedirectPath != "" {
		return auths.DefaultRedirectPath, nil
	}

	return "", errors.ErrNoRedirectPathFound
}
