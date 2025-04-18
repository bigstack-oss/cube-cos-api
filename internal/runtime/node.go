package runtime

import (
	"fmt"
	"os"
	"time"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/node"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func initNodeIdentities() error {
	var err error
	v1.Hostname, err = getHostname()
	if err != nil {
		log.Errorf("runtime: failed to get hostname: %s", err.Error())
		return err
	}

	v1.HostID, err = v1.GenerateNodeHashByMacAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate host id: %s", err.Error())
		return err
	}

	v1.CurrentRole, err = cubecos.GetNodeRole()
	if err != nil {
		log.Errorf("runtime: failed to get node role: %s", err.Error())
		return err
	}

	v1.IsHaEnabled, err = cubecos.IsHaEnabled()
	if err != nil {
		log.Errorf("runtime: failed to get ha enabled: %s", err.Error())
		return err
	}

	v1.MgmtNet, err = cubecos.GetMgmtNet()
	if err != nil {
		log.Errorf("runtime: failed to get management network: %s", err.Error())
		return err
	}

	v1.MgmtIP, err = cubecos.GetManagementIp(v1.MgmtNet)
	if err != nil {
		log.Errorf("runtime: failed to get management ip: %s", err.Error())
		return err
	}

	v1.StorageNet, err = cubecos.GetStorageNet()
	if err != nil {
		log.Errorf("runtime: failed to get storage network: %s", err.Error())
		return err
	}

	v1.StorageIP, err = cubecos.GetStorageIp(v1.StorageNet)
	if err != nil {
		log.Errorf("runtime: failed to get storage ip: %s", err.Error())
		return err
	}

	v1.DataCenterVip, err = cubecos.GetControllerVirtualIp(v1.MgmtNet)
	if err != nil {
		log.Errorf("runtime: failed to get controller virtual ip: %s", err.Error())
		return err
	}

	v1.DataCenterName, err = cubecos.GetDataCenterName()
	if err != nil {
		log.Errorf("runtime: failed to get data center name: %s", err.Error())
		return err
	}

	v1.DataCenterVersion, err = cubecos.GetDataCenterVersion()
	if err != nil {
		log.Errorf("runtime: failed to get data center version: %s", err.Error())
		return err
	}

	v1.DataCenterNumericVersion, err = cubecos.GetDataCenterNumericVersion()
	if err != nil {
		log.Errorf("runtime: failed to get data center numeric version: %s", err.Error())
		return err
	}

	v1.SerialNumber, err = v1.GetSystemSerial()
	if err != nil {
		log.Errorf("runtime: failed to get system serial: %s", err.Error())
	}

	v1.ListenIp = conf.Opts.Spec.Listen.Local
	v1.ListenPort = conf.Opts.Spec.Listen.Port
	v1.ListenAddr = genLocalAddr()
	v1.AdvertisePort = conf.Opts.Spec.Listen.Port
	v1.AdvertiseAddr = genServiceDiscoveryAddr()
	v1.IsGpuEnabled = cubecos.IsGpuEnabled()
	v1.LogoutRedirectUrl = genLogoutRedirectUrl()
	v1.LocalTimeZone = getLocalTimeZone()
	v1.LocalTimeZoneSeconds = getLocalTimeZoneSeconds()
	v1.LocalTimeFixedZone = time.FixedZone("", v1.LocalTimeZoneSeconds)
	cubecos.SyncTunings()
	cubecos.SyncSupportFiles()

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
		"role":         v1.CurrentRole,
		"hostname":     v1.Hostname,
		"dataCenter":   v1.DataCenterName,
		"nodeID":       v1.HostID,
		"serialNumber": v1.SerialNumber,
		"protocol":     conf.Opts.Kind,
		"ip":           v1.MgmtIP,
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
		v1.MgmtIP,
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
