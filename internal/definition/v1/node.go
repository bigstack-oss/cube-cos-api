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
	HostID         string
	Hostname       string
	DataCenterName string
	DataCenterVip  string
	ListenAddr     string
	ListenPort     int
	AdvertiseAddr  string
	AdvertisePort  int
	MgmtNet        string
	MgmtIP         string
	IsHaEnabled    bool
	IsGpuEnabled   bool
)

type Node struct {
	Id           string `json:"id" yaml:"id"`
	Hostname     string `json:"hostname" yaml:"hostname"`
	Role         string `json:"role" yaml:"role"`
	Protocol     string `json:"protocol,omitempty" yaml:"protocol,omitempty" bson:"protocol,omitempty"`
	Address      string `json:"address" yaml:"address"`
	ManagementIP string `json:"managementIP" yaml:"managementIP"`
	License      `json:"license,omitempty" yaml:"license,omitempty" bson:"license,omitempty"`
	Status       string            `json:"status" yaml:"status"`
	Vcpu         ComputeStatistic  `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
	Memory       SpaceStatistic    `json:"memory" yaml:"memory" bson:"memory"`
	Storage      SpaceStatistic    `json:"storage" yaml:"storage" bson:"storage"`
	Uptime       string            `json:"uptime" yaml:"uptime"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" bson:"labels,omitempty"`
}

func GenerateNodeHashByMacAddr() (string, error) {
	macAddr, err := GetMacAddr(NetMajorInterface)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(macAddr))
	return hex.EncodeToString(hash[:])[:8], nil
}

func GetNodesByRole(role string) ([]*Node, error) {
	svcs, err := registry.GetService(role)
	if err != nil {
		log.Errorf("failed to get service %s (%s)", role, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		return nil, nil
	}

	nodes := []*Node{}
	for _, svc := range svcs {
		nodes = append(nodes, getNodesByService(svc, svc.Name)...)
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
