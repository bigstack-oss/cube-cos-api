package cubecos

import (
	"fmt"
	"os/exec"

	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

type NetworkInterface struct {
	Interface   string `json:"dev" yaml:"dev" bson:"dev"`
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busid" yaml:"busid" bson:"busid"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

func GetControllerVirtualIp(net string) (string, error) {
	if base.IsDomainNameEnabled {
		return conf.Opts.Spec.Listen.DomainName, nil
	}

	if !base.IsHaEnabled {
		return GetStandaloneVirtualIp(net)
	}

	return GetClusterVirtualIp()
}

func GetStandaloneVirtualIp(net string) (string, error) {
	if net == "" {
		return "", fmt.Errorf("%s network is empty", net)
	}

	netIfAddrIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, net)
	return GetTuningValue(netIfAddrIp)
}

func GetClusterVirtualIp() (string, error) {
	switch base.CurrentRole {
	case nodes.RoleControl, nodes.RoleControlConverged, nodes.RoleEdgeCore, nodes.RoleModerator:
		return GetTuningValue(CubeSysControllerVip)
	case nodes.RoleCompute, nodes.RoleStorage:
		return GetTuningValue(CubeSysControllerIp)
	}

	return "", fmt.Errorf(
		"unsupported role for reading cluster virtual ip: %s",
		base.CurrentRole,
	)
}

func GetManagementNet() (string, error) {
	return GetTuningValue(CubeSysManagementNetwork)
}

func GetManagementIp(mgmtNet string) (string, error) {
	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return GetTuningValue(netIfAddrMgmtIp)
}

func GetStorageNet() (string, error) {
	return GetTuningValue(CubeSysStorageNetwork)
}

func GetStorageIp(storageNet string) (string, error) {
	if storageNet == "" {
		return "", fmt.Errorf("storage network is empty")
	}

	netIfAddrStorageIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, storageNet)
	return GetTuningValue(netIfAddrStorageIp)
}

func DumpInterfaces() ([]NetworkInterface, error) {
	out, err := exec.Command("hex_sdk", "-v", "-f", "json", "DumpInterface").CombinedOutput()
	if err != nil {
		log.Errorf("net: failed to get network info(%v)", err)
		return nil, err
	}

	interfaces := []NetworkInterface{}
	err = json.Unmarshal(out, &interfaces)
	if err != nil {
		log.Errorf("net: failed to unmarshal network info(%v)", err)
		return nil, err
	}

	return interfaces, nil
}

func IsOvnSFlowEnabled() bool {
	_, err := exec.Command("hex_sdk", "ovn_sflow_status").Output()
	if err == nil {
		return true
	}

	result, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}

	return result.ExitCode() == 0
}
