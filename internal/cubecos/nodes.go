package cubecos

import (
	"encoding/json"
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

const (
	EtcNodeInventory = "/etc/settings.cluster.json"
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
