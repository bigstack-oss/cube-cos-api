package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

const (
	Nodes             = "nodes"
	DataCenterHelpUrl = "https://www.bigstack.co/contact-us"
)

var (
	HostID                   string
	Hostname                 string
	DataCenterName           string
	DataCenterVersion        string
	DataCenterNumericVersion string
	DataCenterVip            string
	ListenIp                 string
	ListenAddr               string
	ListenPort               int
	AdvertiseIp              string
	AdvertiseAddr            string
	AdvertisePort            int
	MgmtNet                  string
	MgmtIP                   string
	IsHaEnabled              bool
	IsGpuEnabled             bool
)

type Node struct {
	Id                string `json:"id" yaml:"id"`
	DataCenter        string `json:"dataCenter" yaml:"dataCenter"`
	Hostname          string `json:"hostname" yaml:"hostname"`
	Role              string `json:"role" yaml:"role"`
	Protocol          string `json:"protocol,omitempty" yaml:"protocol,omitempty" bson:"protocol,omitempty"`
	Address           string `json:"address" yaml:"address"`
	Ip                string `json:"ip" yaml:"ip"`
	ManagementIP      string `json:"managementIP" yaml:"managementIP"`
	License           `json:"license" yaml:"license,omitempty" bson:"license,omitempty"`
	Status            string             `json:"status" yaml:"status"`
	CpuSpec           string             `json:"cpuSpec" yaml:"cpuSpec" bson:"cpuSpec"`
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces" yaml:"networkInterfaces" bson:"networkInterfaces"`
	BlockDevices      []BlockDevice      `json:"blockDevices" yaml:"blockDevices" bson:"blockDevices"`
	Vcpu              ComputeStatistic   `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
	Memory            SpaceStatistic     `json:"memory" yaml:"memory" bson:"memory"`
	Storage           SpaceStatistic     `json:"storage" yaml:"storage" bson:"storage"`
	UptimeSeconds     float64            `json:"uptimeSeconds" yaml:"uptimeSeconds" bson:"uptimeSeconds"`
	Token             string             `json:"-" yaml:"-" bson:"-"`
	Labels            map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty" bson:"labels,omitempty"`
}

type NetworkInterface struct {
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busIdSlaves" yaml:"busIdSlaves" bson:"busIdSlaves"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

type BlockDevice struct {
	Name    string  `json:"device" yaml:"device" bson:"device"`
	Type    string  `json:"type" yaml:"type" bson:"type"`
	SizeMiB float64 `json:"sizeMiB" yaml:"sizeMiB" bson:"sizeMiB"`
	Status  string  `json:"status" yaml:"status" bson:"status"`
}

type RawBlockDevice struct {
	Name        string   `json:"name"`
	MajMin      string   `json:"maj:min"`
	Rm          bool     `json:"rm"`
	Size        string   `json:"size"`
	Ro          bool     `json:"ro"`
	Type        string   `json:"type"`
	MountPoints []string `json:"mountpoints"`
}

func (n *Node) GetBearerToken() string {
	return fmt.Sprintf("Bearer %s", n.Token)
}

func (n *Node) GenAuthHeader() (string, string) {
	return "Authorization", n.GetBearerToken()
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

func (n *Node) GetNodeDetailsUrl() string {
	u := url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
		Path: fmt.Sprintf(
			"/api/v1/datacenters/%s/nodes/%s",
			n.DataCenter,
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

func (n *Node) GetSupportFileUrl() string {
	u := url.URL{
		Scheme:   n.Protocol,
		Host:     n.Address,
		Path:     fmt.Sprintf("/api/v1/datacenters/%s/tunings", n.DataCenter),
		RawQuery: fmt.Sprintf("host=%s", n.Hostname),
	}

	return u.String()
}

func (n *Node) PatchTuningUrl(tuning Tuning) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters/%s", DataCenterName, tuning.Name)
	return u.String()
}

func (n *Node) PatchTuningTaskUrl(tuning Tuning) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/tasks/%s", DataCenterName, tuning.Id)
	return u.String()
}

func (n *Node) PatchTriggerTaskUrl(trigger trigger.Options) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/tasks/%s", DataCenterName, trigger.Id)
	return u.String()
}

func (n *Node) CreateSupportFileUrl(supportFile SupportFile) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles", DataCenterName)
	return u.String()
}

func (n *Node) PatchSupportFileTaskUrl(supportFile SupportFile) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/tasks/%s", DataCenterName, supportFile.Id)
	return u.String()
}

func (n *Node) IsLocal() bool {
	return n.Address == AdvertiseAddr && n.Role == CurrentRole
}

func IsLocalNode(hostname string) bool {
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

func GetOneOfControllerNode() (*Node, error) {
	nodes, err := GetControllerNodes()
	if err != nil {
		return nil, err
	}

	return nodes[0], nil
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
