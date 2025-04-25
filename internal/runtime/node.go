package runtime

import (
	"fmt"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/errors"
	log "go-micro.dev/v5/logger"
)

func initIdentities() error {
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

	v1.ManagementNet, err = cubecos.GetManagementNet()
	if err != nil {
		log.Errorf("runtime: failed to get management network: %s", err.Error())
		return err
	}

	v1.ManagementIp, err = cubecos.GetManagementIp(v1.ManagementNet)
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

	v1.DataCenterVip, err = cubecos.GetControllerVirtualIp(v1.ManagementNet)
	if err != nil {
		log.Errorf("runtime: failed to get controller virtual ip: %s", err.Error())
		return err
	}

	v1.DataCenterName, err = cubecos.GetDataCenterName()
	if err != nil {
		log.Errorf("runtime: failed to get data center name: %s", err.Error())
		return err
	}

	v1.ServiceDiscoveryIdentity, err = parseServiceDiscoveryIdentity()
	if err != nil {
		log.Errorf("runtime: failed to parse service discovery identify: %s", err.Error())
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

	v1.SerialNumber, err = v1.GetSystemSerial(conf.Opts.Identity.Serial)
	if err != nil {
		log.Warnf("runtime: failed to get system serial: %s", err.Error())
	}

	v1.ListenIp, err = parseLocalListenAddr()
	if err != nil {
		log.Errorf("runtime: failed to parse local listen address: %s", err.Error())
		return err
	}

	v1.ListenPort, err = parseLocalListenPort()
	if err != nil {
		log.Errorf("runtime: failed to parse local listen port: %s", err.Error())
		return err
	}

	v1.ListenAddr, err = genLocalAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate local address: %s", err.Error())
		return err
	}

	v1.AdvertisePort, err = parseAdvertisePort()
	if err != nil {
		log.Errorf("runtime: failed to parse advertise port: %s", err.Error())
		return err
	}

	v1.AdvertiseAddr, err = genServiceDiscoveryAddr()
	if err != nil {
		log.Errorf("runtime: failed to generate advertise address: %s", err.Error())
		return err
	}

	v1.IsGpuEnabled, err = cubecos.IsGpuEnabled()
	if err != nil {
		log.Warnf("runtime: failed to get gpu enablement: %s", err.Error())
	}

	v1.LogoutRedirectUrl, err = genLogoutRedirectUrl()
	if err != nil {
		log.Errorf("runtime: failed to generate logout redirect url: %s", err.Error())
		return err
	}

	return nil
}

func parseServiceDiscoveryIdentity() (string, error) {
	if v1.DataCenterName == "" {
		return "", errors.InvalidDataCenterName
	}

	if v1.DataCenterVip == "" {
		return "", errors.InvalidListenAddress
	}

	return fmt.Sprintf(
		"%s-%s",
		v1.DataCenterName,
		v1.DataCenterVip,
	), nil
}

func parseLocalListenAddr() (string, error) {
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
