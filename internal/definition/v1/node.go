package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"

	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

const (
	Nodes             = "nodes"
	DataCenterHelpUrl = "https://www.bigstack.co/contact-us"
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
	DataCenter    string `json:"dataCenter" yaml:"dataCenter"`
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
	Token         string            `json:"-" yaml:"-" bson:"-"`
	Labels        map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" bson:"labels,omitempty"`
}

func (n *Node) GetBearerToken() string {
	return fmt.Sprintf("Bearer %s", n.Token)
}

func (n *Node) GetMetricUrl(metric, view string) string {
	u := url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
		Path: fmt.Sprintf(
			"/api/v1/datacenters/%s/metrics/%s/%s/hosts/%s",
			n.DataCenter,
			metric,
			view,
			n.Hostname,
		),
	}

	return u.String()
}

func (n *Node) GetTuningUrl() string {
	u := url.URL{
		Scheme:   n.Protocol,
		Host:     n.Address,
		Path:     fmt.Sprintf("/api/v1/datacenters/%s/tunings", n.DataCenter),
		RawQuery: "allNodes=false",
	}

	return u.String()
}

func IsCurrentHost(hostname string) bool {
	return Hostname == hostname
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

func GetNodeByHostname(hostname string) (*Node, error) {
	nodes, err := ListNodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.Hostname == hostname {
			return node, nil
		}
	}

	return nil, fmt.Errorf(
		"failed to get node by hostname %s",
		hostname,
	)
}
