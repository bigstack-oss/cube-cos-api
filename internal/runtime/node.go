package runtime

import (
	"fmt"
	"os"
	"time"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/node"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func initNodeIdentities() error {
	var err error
	definition.Hostname, err = getHostname()
	if err != nil {
		log.Errorf("failed to get hostname: %s", err.Error())
		return err
	}

	definition.HostID, err = definition.GenerateNodeHashByMacAddr()
	if err != nil {
		log.Errorf("failed to generate host id: %s", err.Error())
		return err
	}

	definition.CurrentRole, err = cubecos.GetNodeRole()
	if err != nil {
		log.Errorf("failed to get node role: %s", err.Error())
		return err
	}

	definition.IsHaEnabled, err = cubecos.IsHaEnabled()
	if err != nil {
		log.Errorf("failed to get ha enabled: %s", err.Error())
		return err
	}

	definition.MgmtNet, err = cubecos.GetMgmtNet()
	if err != nil {
		log.Errorf("failed to get management network: %s", err.Error())
		return err
	}

	definition.MgmtIP, err = cubecos.GetManagementIp(definition.MgmtNet)
	if err != nil {
		log.Errorf("failed to get management ip: %s", err.Error())
		return err
	}

	definition.DataCenterVip, err = cubecos.GetControllerVirtualIp(definition.MgmtNet)
	if err != nil {
		log.Errorf("failed to get controller virtual ip: %s", err.Error())
		return err
	}

	definition.DataCenterName, err = cubecos.GetDataCenterName()
	if err != nil {
		log.Errorf("failed to get data center name: %s", err.Error())
		return err
	}

	definition.DataCenterVersion, err = cubecos.GetDataCenterVersion()
	if err != nil {
		log.Errorf("failed to get data center version: %s", err.Error())
		return err
	}

	definition.ListenIp = conf.Opts.Spec.Listen.Local
	definition.ListenPort = conf.Opts.Spec.Listen.Port
	definition.ListenAddr = genLocalAddr()
	definition.AdvertisePort = conf.Opts.Spec.Listen.Port
	definition.AdvertiseAddr = genServiceDiscoveryAddr()
	definition.IsGpuEnabled = cubecos.IsGpuEnabled()
	definition.LogoutRedirectUrl = genLogoutRedirectUrl()
	definition.LocalTimeZone = getLocalTimeZone()
	definition.LocalTimeZoneSeconds = getLocalTimeZoneSeconds()
	definition.LocalTimeFixedZone = time.FixedZone("", definition.LocalTimeZoneSeconds)
	cubecos.SyncTunings()

	return nil
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

func initNodePeerSyncer() {
	service.RegisterOperator(node.Name(), &node.Operator{})
}

func genNodeMetadata() map[string]string {
	return map[string]string{
		"role":         definition.CurrentRole,
		"hostname":     definition.Hostname,
		"dataCenter":   definition.DataCenterName,
		"nodeID":       definition.HostID,
		"protocol":     conf.Opts.Kind,
		"ip":           definition.MgmtIP,
		"isGpuEnabled": fmt.Sprintf("%t", definition.IsGpuEnabled),
		"token":        definition.DefaultNodeToken,
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
		definition.MgmtIP,
		conf.Opts.Spec.Listen.Port,
	)
}

func genLogoutRedirectUrl() string {
	return fmt.Sprintf(
		"https://%s:4443%s",
		definition.DataCenterVip,
		conf.Opts.Spec.Identity.LogoutRedirect,
	)
}
