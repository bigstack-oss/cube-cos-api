package cubecos

import (
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

type NetworkInterface struct {
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busid" yaml:"busid" bson:"busid"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

func GetControllerVirtualIp(mgmtNet string) (string, error) {
	if !base.IsHaEnabled {
		return GetStandaloneVirtualIp(mgmtNet)
	}

	return GetClusterVirtualIp()
}

func GetStandaloneVirtualIp(mgmtNet string) (string, error) {
	if mgmtNet == "" {
		return "", fmt.Errorf("management network is empty")
	}

	netIfAddrMgmtIp := fmt.Sprintf("%s%s", CubeNetIfAddrPrefix, mgmtNet)
	return GetTuningValue(netIfAddrMgmtIp)
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
	out, err := exec.Command("hex_sdk", "-f", "json", "DumpInterface").CombinedOutput()
	if err != nil {
		log.Errorf("net: failed to get network info: %s", err.Error())
		return nil, err
	}

	interfaces := []NetworkInterface{}
	err = json.Unmarshal(out, &interfaces)
	if err != nil {
		log.Errorf("net: failed to unmarshal network info: %s", err.Error())
		return nil, err
	}

	return interfaces, nil
}
