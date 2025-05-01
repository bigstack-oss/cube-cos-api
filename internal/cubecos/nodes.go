package cubecos

import (
	"encoding/json"
	"os"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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

func GetSourceNodeMap() (map[string]v1.Node, error) {
	file, err := os.Open(EtcNodeInventory)
	if err != nil {
		log.Errorf("nodes: failed to open %s: %v", EtcNodeInventory, err)
		return nil, err
	}

	defer file.Close()
	nodes := map[string]Node{}
	err = json.NewDecoder(file).Decode(&nodes)
	if err != nil {
		return nil, err
	}

	apiNodes := map[string]v1.Node{}
	for _, node := range nodes {
		apiNodes[node.Hostname] = v1.Node{
			Hostname:     node.Hostname,
			Role:         node.Role,
			Ip:           node.Ip.Provider,
			ManagementIP: node.Ip.Management,
			StorageIP:    node.Ip.Storage,
		}
	}

	return apiNodes, nil
}
