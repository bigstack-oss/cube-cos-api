package cubecos

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1/accelerators/devices"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
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
		log.Errorf("nodes: failed to open %s(%v)", EtcNodeInventory, err)
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

func ListNodesWithTimeSensitiveInfo() []nodes.Node {
	list := nodes.List()
	if len(list) == 0 {
		return []nodes.Node{}
	}

	syncTimeSensitiveInfo(&list)
	return list
}

func GetNodeWithTimeSensitiveInfo(hostname string) (*nodes.Node, error) {
	for _, node := range ListNodesWithTimeSensitiveInfo() {
		if node.Hostname == hostname {
			return &node, nil
		}
	}

	return nil, fmt.Errorf(
		"node %s not found",
		hostname,
	)
}

func IsGpuEnabled() (bool, error) {
	opts := conf.GetOpenstack()
	provider, err := openstack.NewProvider(opts.Auth.File)
	if err != nil {
		log.Errorf("gpu: failed to create openstack provider(%v)", err)
		return false, err
	}

	accelerator, err := openstack.NewAcceleratorV1(
		provider,
		openstack.DefaultEndpointOpts,
	)
	if err != nil {
		log.Errorf("gpu: failed to create accelerator client(%v)", err)
		return false, err
	}

	devices, err := devices.List(
		accelerator,
		devices.ListOpts{Hostname: base.Hostname},
	)
	if err != nil {
		log.Errorf("gpu: failed to list accelerator devices(%v)", err)
		return false, err
	}

	return len(devices) > 0, nil
}

func syncTimeSensitiveInfo(list *[]nodes.Node) {
	syncLicense(list)
	syncVirutalIpOwner(list)
	syncPowerStatus(list)
	syncIpmiInfo(list)
}

func syncLicense(list *[]nodes.Node) {
	for i, node := range *list {
		(*list)[i].License = GetHostLicense(node.Hostname)
	}
}

func syncVirutalIpOwner(list *[]nodes.Node) {
	for i, node := range *list {
		(*list)[i].IsVirtualIpOwner = pacemaker.IsVirtualIpOwner(node.Hostname)
	}
}

func syncPowerStatus(list *[]nodes.Node) {
	for i, node := range *list {
		if !hasPowerRequest(node.Hostname) {
			continue
		}

		status, err := getNodePendingPowerStatus(node.Status)
		if err != nil {
			continue
		}

		(*list)[i].Status = status
	}
}

func syncIpmiInfo(list *[]nodes.Node) {
	for i, node := range *list {
		if hasIpmiRecord(node.Hostname) {
			(*list)[i].IsIpmiConnected = true
		}
	}
}

func hasPowerRequest(hostname string) bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		nodes.Db,
		nodes.RequestsCollection,
		bson.M{"hostname": hostname},
	)
	if err != nil {
		log.Errorf("nodes: failed to get node(%s) power request count(%v)", hostname, err)
		return false
	}

	return count > 1
}

func hasIpmiRecord(hostname string) bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		nodes.Db,
		nodes.CollectionIpmi,
		bson.M{"host": hostname},
	)
	if err != nil {
		log.Errorf("nodes: failed to get node(%s) ipmi record(%v)", hostname, err)
		return false
	}

	return count > 0
}

func getNodePendingPowerStatus(hostname string) (string, error) {
	mongo := mongo.GetGlobalHelper()
	doc, err := mongo.Get(
		nodes.Db,
		nodes.RequestsCollection,
		bson.M{"hostname": hostname},
	)
	if err != nil {
		log.Errorf("nodes: failed to get node(%s) status(%v)", hostname, err)
		return "", err
	}
	if doc == nil {
		return "", err
	}

	node := &nodes.Node{}
	err = doc.Decode(node)
	if err != nil {
		log.Errorf("nodes: failed to decode node(%s) status(%v)", hostname, err)
		return "", err
	}

	return node.Status, nil
}
