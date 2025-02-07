package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

const (
	Nodes = "nodes"
)

var (
	HostID            string
	Hostname          string
	DataCenterName    string
	DataCenterVersion string
	DataCenterVip     string
	ListenAddr        string
	ListenPort        int
	AdvertiseAddr     string
	AdvertisePort     int
	MgmtNet           string
	MgmtIP            string
	IsHaEnabled       bool
	IsGpuEnabled      bool
)

type Node struct {
	Id            string `json:"id" yaml:"id"`
	Hostname      string `json:"hostname" yaml:"hostname"`
	Role          string `json:"role" yaml:"role"`
	Protocol      string `json:"protocol,omitempty" yaml:"protocol,omitempty" bson:"protocol,omitempty"`
	Address       string `json:"address" yaml:"address"`
	ManagementIP  string `json:"managementIP" yaml:"managementIP"`
	License       `json:"license,omitempty" yaml:"license,omitempty" bson:"license,omitempty"`
	Status        string            `json:"status" yaml:"status"`
	Vcpu          ComputeStatistic  `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
	Memory        SpaceStatistic    `json:"memory" yaml:"memory" bson:"memory"`
	Storage       SpaceStatistic    `json:"storage" yaml:"storage" bson:"storage"`
	UptimeSeconds float64           `json:"uptimeSeconds" yaml:"uptimeSeconds" bson:"uptimeSeconds"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" bson:"labels,omitempty"`
}

func GenerateNodeHashByMacAddr() (string, error) {
	macAddr, err := GetMacAddr(NetMajorInterface)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(macAddr))
	return hex.EncodeToString(hash[:])[:8], nil
}

func GetNodesByRole(roleName string) ([]*Node, error) {
	svcs, err := registry.GetService(DataCenterName)
	if err != nil {
		log.Errorf("failed to get %s role from service %s (%s)", roleName, DataCenterName, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		return nil, nil
	}

	nodes := []*Node{}
	for _, svc := range svcs {
		roleNodes := parseNodesByRole(svc, roleName)
		if len(roleNodes) == 0 {
			continue
		}

		nodes = append(nodes, roleNodes...)
	}

	return nodes, nil
}

func ListNodes() ([]*Node, error) {
	svcs, err := registry.GetService(DataCenterName)
	if err != nil {
		log.Errorf("failed to get nodes from %s (%s)", DataCenterName, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		return nil, nil
	}

	nodes := []*Node{}
	for _, svc := range svcs {
		nodes = append(nodes, parseNodes(svc)...)
	}

	return nodes, nil
}

func GetControllerNodes() ([]*Node, error) {
	nodes, err := GetNodesByRole("control")
	if err == nil && len(nodes) > 0 {
		return nodes, nil
	}

	nodes, err = GetNodesByRole("control-converged")
	if err == nil && len(nodes) > 0 {
		return nodes, nil
	}

	return nil, fmt.Errorf(
		"failed to get control nodes(control or control-converged): %s",
		err.Error(),
	)
}
