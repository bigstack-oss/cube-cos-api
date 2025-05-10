package runtime

import (
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	log "go-micro.dev/v5/logger"
)

func initIdentities() error {
	var err error
	base.Hostname, err = getHostname()
	if err != nil {
		log.Errorf("runtime: failed to get hostname: %s", err.Error())
		return err
	}

	base.HostID, err = base.GenerateNodeHashByMacAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate host id: %s", err.Error())
		return err
	}

	base.CurrentRole, err = cubecos.GetNodeRole()
	if err != nil {
		log.Errorf("runtime: failed to get node role: %s", err.Error())
		return err
	}

	base.IsHaEnabled, err = cubecos.IsHaEnabled()
	if err != nil {
		log.Errorf("runtime: failed to get ha enabled: %s", err.Error())
		return err
	}

	base.ManagementNet, err = cubecos.GetManagementNet()
	if err != nil {
		log.Errorf("runtime: failed to get management network: %s", err.Error())
		return err
	}

	base.ManagementIp, err = cubecos.GetManagementIp(base.ManagementNet)
	if err != nil {
		log.Errorf("runtime: failed to get management ip: %s", err.Error())
		return err
	}

	base.StorageNet, err = cubecos.GetStorageNet()
	if err != nil {
		log.Errorf("runtime: failed to get storage network: %s", err.Error())
		return err
	}

	base.StorageIP, err = cubecos.GetStorageIp(base.StorageNet)
	if err != nil {
		log.Errorf("runtime: failed to get storage ip: %s", err.Error())
		return err
	}

	base.DataCenterVip, err = cubecos.GetControllerVirtualIp(base.ManagementNet)
	if err != nil {
		log.Errorf("runtime: failed to get controller virtual ip: %s", err.Error())
		return err
	}

	base.DataCenterName, err = cubecos.GetDataCenterName()
	if err != nil {
		log.Errorf("runtime: failed to get data center name: %s", err.Error())
		return err
	}

	base.DataCenterVersion, err = cubecos.GetDataCenterVersion()
	if err != nil {
		log.Errorf("runtime: failed to get data center version: %s", err.Error())
		return err
	}

	base.DataCenterNumericVersion, err = cubecos.GetDataCenterNumericVersion()
	if err != nil {
		log.Errorf("runtime: failed to get data center numeric version: %s", err.Error())
		return err
	}

	base.SerialNumber, err = base.GetSystemSerial(conf.Opts.Identity.Serial)
	if err != nil {
		log.Warnf("runtime: failed to get system serial: %s", err.Error())
	}

	base.ListenIp, err = parseLocalListenAddr()
	if err != nil {
		log.Errorf("runtime: failed to parse local listen address: %s", err.Error())
		return err
	}

	base.ListenPort, err = parseLocalListenPort()
	if err != nil {
		log.Errorf("runtime: failed to parse local listen port: %s", err.Error())
		return err
	}

	base.ListenAddr, err = genLocalAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate local address: %s", err.Error())
		return err
	}

	base.AdvertisePort, err = parseAdvertisePort()
	if err != nil {
		log.Errorf("runtime: failed to parse advertise port: %s", err.Error())
		return err
	}

	base.AdvertiseAddr, err = genServiceDiscoveryAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate advertise address: %s", err.Error())
		return err
	}

	base.IsGpuEnabled, err = cubecos.IsGpuEnabled()
	if err != nil {
		log.Warnf("runtime: failed to get gpu enablement: %s", err.Error())
	}

	auths.RedirectPath, err = parseRedirectPath()
	if err != nil {
		log.Errorf("runtime: failed to parse redirect path: %s", err.Error())
	}

	auths.RedirectUrl, err = genLogoutRedirectUrl()
	if err != nil {
		log.Errorf("runtime: failed to generate logout redirect url: %s", err.Error())
		return err
	}

	return nil
}
