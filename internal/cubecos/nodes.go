package cubecos

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1/accelerators/devices"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

const (
	EtcNodeInventory = "/etc/settings.cluster.json"
	cubeSysRole      = "cubesys.role"
)

type Node struct {
	Hostname string `json:"hostname"`
	Role     string `json:"role"`
	Ip       `json:"ip"`
}

type Ip struct {
	Management string `json:"management"`
	Provider   string `json:"provider"`
	Overlay    string `json:"overlay"`
	Storage    string `json:"storage"`
}

func GetSourceNodeMap() (map[string]nodes.Node, error) {
	file, err := os.Open(EtcNodeInventory)
	if err != nil {
		log.Errorf("nodes: failed to open %s: %v", EtcNodeInventory, err)
		return nil, err
	}

	defer file.Close()
	srcNodes := map[string]Node{}
	err = json.NewDecoder(file).Decode(&srcNodes)
	if err != nil {
		return nil, err
	}

	nodeMap := map[string]nodes.Node{}
	for _, srcNode := range srcNodes {
		nodeMap[srcNode.Hostname] = nodes.Node{
			Hostname:     srcNode.Hostname,
			Role:         srcNode.Role,
			Ip:           srcNode.Ip.Provider,
			ManagementIP: srcNode.Ip.Management,
			StorageIP:    srcNode.Ip.Storage,
		}
	}

	return nodeMap, nil
}

func GetNodeRole() (string, error) {
	role, err := GetTuningValue(cubeSysRole)
	if err != nil {
		return "", err
	}

	if role == "" {
		return "", fmt.Errorf("role is empty")
	}

	return role, nil
}

func IsGpuEnabled() (bool, error) {
	opts := conf.GetOpenstack()
	provider, err := openstack.NewProvider(opts.Auth.File)
	if err != nil {
		log.Errorf("gpu: failed to create openstack provider: %v", err)
		return false, err
	}

	accelerator, err := openstack.NewAcceleratorV1(
		provider,
		openstack.DefaultEndpointOpts,
	)
	if err != nil {
		log.Errorf("gpu: failed to create accelerator client: %v", err)
		return false, err
	}

	devices, err := devices.List(
		accelerator,
		devices.ListOpts{Hostname: base.Hostname},
	)
	if err != nil {
		log.Errorf("gpu: failed to list accelerator devices: %v", err)
		return false, err
	}

	return len(devices) > 0, nil
}
