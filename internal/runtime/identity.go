package runtime

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	log "go-micro.dev/v5/logger"
)

func initIdentities() error {
	var err error
	base.SystemSeed, err = cubecos.GetSystemSeed()
	if err != nil {
		log.Errorf("runtime: failed to get system seed(%v)", err)
		return err
	}

	base.Hostname, err = getHostname()
	if err != nil {
		log.Errorf("runtime: failed to get hostname(%v)", err)
		return err
	}

	base.HostID, err = base.GenerateNodeHashByMacAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate host id(%v)", err)
		return err
	}

	base.CurrentRole, err = cubecos.GetNodeRole()
	if err != nil {
		log.Errorf("runtime: failed to get node role(%v)", err)
		return err
	}

	base.IsHaEnabled, err = cubecos.IsHaEnabled()
	if err != nil {
		log.Errorf("runtime: failed to get ha enabled(%v)", err)
		return err
	}

	base.ManagementNet, err = cubecos.GetManagementNet()
	if err != nil {
		log.Errorf("runtime: failed to get management network(%v)", err)
		return err
	}

	base.ManagementIp, err = cubecos.GetManagementIp(base.ManagementNet)
	if err != nil {
		log.Errorf("runtime: failed to get management ip(%v)", err)
		return err
	}

	base.StorageNet, err = cubecos.GetStorageNet()
	if err != nil {
		log.Errorf("runtime: failed to get storage network(%v)", err)
		return err
	}

	base.StorageIP, err = cubecos.GetStorageIp(base.StorageNet)
	if err != nil {
		log.Errorf("runtime: failed to get storage ip(%v)", err)
		return err
	}

	base.DataCenterVip, err = cubecos.GetControllerVirtualIp(base.ManagementNet)
	if err != nil {
		log.Errorf("runtime: failed to get controller virtual ip(%v)", err)
		return err
	}

	base.DataCenterName, err = cubecos.GetDataCenterName()
	if err != nil {
		log.Errorf("runtime: failed to get data center name(%v)", err)
		return err
	}

	base.DataCenterVersion, err = cubecos.GetDataCenterVersion()
	if err != nil {
		log.Errorf("runtime: failed to get data center version(%v)", err)
		return err
	}

	base.DataCenterNumericVersion, err = cubecos.GetDataCenterNumericVersion()
	if err != nil {
		log.Errorf("runtime: failed to get data center numeric version(%v)", err)
		return err
	}

	base.SerialNumber, err = cubecos.GetSystemSerial()
	if err != nil {
		log.Warnf("runtime: failed to get system serial(%v)", err)
	}

	base.ListenIp, err = parseLocalListenAddr()
	if err != nil {
		log.Errorf("runtime: failed to parse local listen address(%v)", err)
		return err
	}

	base.ListenPort, err = parseLocalListenPort()
	if err != nil {
		log.Errorf("runtime: failed to parse local listen port(%v)", err)
		return err
	}

	base.ListenAddr, err = genLocalAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate local address(%v)", err)
		return err
	}

	base.AdvertisePort, err = parseAdvertisePort()
	if err != nil {
		log.Errorf("runtime: failed to parse advertise port(%v)", err)
		return err
	}

	base.AdvertiseAddr, err = genServiceDiscoveryAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate advertise address(%v)", err)
		return err
	}

	base.IsGpuEnabled, err = cubecos.IsGpuEnabled()
	if err != nil {
		log.Warnf("runtime: failed to get gpu enablement(%v)", err)
	}

	auths.RedirectPath, err = parseRedirectPath()
	if err != nil {
		log.Errorf("runtime: failed to parse redirect path(%v)", err)
	}

	auths.RedirectUrl, err = genLogoutRedirectUrl()
	if err != nil {
		log.Errorf("runtime: failed to generate logout redirect url(%v)", err)
		return err
	}

	return nil
}
