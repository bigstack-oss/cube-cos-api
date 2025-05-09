package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"slices"
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
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
	ManagementNet            string
	ManagementIp             string
	StorageNet               string
	StorageIP                string
	IsHaEnabled              bool
	IsGpuEnabled             bool
	NodeMetadata             map[string]string

	getRegisteredServices = sync.Mutex{}
	UpdateNodes           = sync.Mutex{}
	nodes                 = []Node{}
)

type Node struct {
	Id                string             `json:"id" yaml:"id"`
	SerialNumber      string             `json:"serialNumber" yaml:"serialNumber"`
	DataCenter        string             `json:"dataCenter" yaml:"dataCenter"`
	Hostname          string             `json:"hostname" yaml:"hostname"`
	Role              string             `json:"role" yaml:"role"`
	Protocol          string             `json:"protocol,omitempty" yaml:"protocol,omitempty" bson:"protocol,omitempty"`
	Address           string             `json:"address" yaml:"address"`
	Ip                string             `json:"ip" yaml:"ip"`
	ManagementIP      string             `json:"managementIP" yaml:"managementIP"`
	StorageIP         string             `json:"storageIP" yaml:"storageIP"`
	License           license.Options    `json:"license" yaml:"license,omitempty" bson:"license,omitempty"`
	Status            string             `json:"status" yaml:"status"`
	CpuSpec           string             `json:"cpuSpec" yaml:"cpuSpec" bson:"cpuSpec"`
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces" yaml:"networkInterfaces" bson:"networkInterfaces"`
	BlockDevices      []BlockDevice      `json:"blockDevices" yaml:"blockDevices" bson:"blockDevices"`
	Vcpu              ComputeStatistic   `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
	Memory            SpaceStatistic     `json:"memory" yaml:"memory" bson:"memory"`
	Storage           SpaceStatistic     `json:"storage" yaml:"storage" bson:"storage"`
	UptimeSeconds     float64            `json:"uptimeSeconds" yaml:"uptimeSeconds" bson:"uptimeSeconds"`
	Labels            map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty" bson:"labels,omitempty"`
}

type NetworkInterface struct {
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busIdSlaves" yaml:"busIdSlaves" bson:"busIdSlaves"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

type RawNetworkInterface struct {
	Label       string `json:"label" yaml:"label" bson:"label"`
	BusIdSlaves string `json:"busid" yaml:"busid" bson:"busid"`
	Driver      string `json:"driver" yaml:"driver" bson:"driver"`
	State       string `json:"state" yaml:"state" bson:"state"`
	Speed       string `json:"speed" yaml:"speed" bson:"speed"`
}

type BlockDevice struct {
	Serial       string             `json:"serial"`
	Name         string             `json:"device" yaml:"device" bson:"device"`
	Type         string             `json:"type" yaml:"type" bson:"type"`
	SizeMiB      float64            `json:"sizeMiB" yaml:"sizeMiB" bson:"sizeMiB"`
	Availability string             `json:"availability" yaml:"availability" bson:"availability"`
	Status       status.BlockDevice `json:"status" yaml:"status" bson:"status"`
}

type RawBlockDevice struct {
	Type        string   `json:"type"`
	Serial      string   `json:"serial"`
	Name        string   `json:"name"`
	Size        string   `json:"size"`
	Rota        bool     `json:"rota"`
	MountPoints []string `json:"mountpoints"`
}

func (r *RawBlockDevice) IsPartition() bool {
	return r.Type == "part"
}

func (r *RawBlockDevice) IsBlock() bool {
	return r.Type == "disk"
}

func (r *RawBlockDevice) NoMountPoints() bool {
	return len(r.MountPoints) == 0
}

func (n *Node) GenUrl() string {
	u := url.URL{Scheme: n.Protocol, Host: n.Address}
	return u.String()
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
		Path:     fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters", n.DataCenter),
		RawQuery: "allNodes=false",
	}

	return u.String()
}

func (n *Node) GetSettingUrl(path string) string {
	u := url.URL{
		Scheme:   n.Protocol,
		Host:     n.Address,
		Path:     path,
		RawQuery: "clusterWise=false",
	}

	return u.String()
}

func (n *Node) GetSupportFileUrl() string {
	u := url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
		Path:   fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/hosts/%s", n.DataCenter, n.Hostname),
	}

	return u.String()
}

func (n *Node) DownloadSupportFileUrl(setname, filename string) string {
	u := url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
		Path:   fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/%s/%s", n.DataCenter, setname, filename),
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

func (n *Node) PatchTriggerTaskUrl(trigger trigger.ApiOptions) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/tasks/%s", DataCenterName, trigger.Name)
	return u.String()
}

func (n *Node) CreateSupportFileUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles", DataCenterName)
	return u.String()
}

func (n *Node) PatchSupportFileTaskUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/%s", DataCenterName, file.Group)
	return u.String()
}

func (n *Node) PatchSettingTaskUrl(setting setting.Options) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/settings/tasks", DataCenterName)
	return u.String()
}

func (n *Node) DeleteRepairingTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/healths/tasks/repairing", DataCenterName)
	return u.String()
}

func (n *Node) DeleteModuleRepairingTaskUrl(module string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/healths/tasks/repairing/%s", DataCenterName, module)
	return u.String()
}

func (n *Node) IsLocal() bool {
	return n.Address == AdvertiseAddr && n.Hostname == Hostname
}

func (n *Node) IsDown() bool {
	return n.Status == status.Down
}

func (n *Node) MatchHardwareSerial(hardwareSerials []string) bool {
	return slices.Contains(hardwareSerials, n.SerialNumber)
}

func (n *Node) IsLicenseExpired() bool {
	return n.License.Status.Current == status.Expired
}

func (n *Node) IsUnlicensed() bool {
	return n.License.Status.Current == status.Unlicense
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

func GetNodesByRole(roleName string) ([]Node, error) {
	svcs, err := GetRegisteredServices()
	if err != nil {
		return nil, err
	}

	nodes := []Node{}
	for _, svc := range svcs {
		roleNodes := parseNodesByRole(svc, roleName)
		if len(roleNodes) == 0 {
			continue
		}

		nodes = append(nodes, roleNodes...)
	}

	log.Infof("-----------------------")

	for _, node := range nodes {
		log.Infof(node.Hostname)
	}

	return nodes, nil
}

func GetRegisteredServices() ([]*registry.Service, error) {
	getRegisteredServices.Lock()
	defer getRegisteredServices.Unlock()

	svcs, err := registry.GetService(ServiceDiscoveryIdentity)
	if err != nil {
		log.Errorf("failed to get service from %s (%s)", ServiceDiscoveryIdentity, err.Error())
		return nil, err
	}

	if len(svcs) <= 0 {
		err := fmt.Errorf("no any service find from %s", ServiceDiscoveryIdentity)
		log.Errorf(err.Error())
		return nil, err
	}

	return svcs, nil
}

func HostnameNodeMap() (map[string]Node, error) {
	svcs, err := GetRegisteredServices()
	if err != nil {
		return nil, err
	}

	nodeMap := map[string]Node{}
	for _, svc := range svcs {
		nodes := parseNodes(svc)
		for _, node := range nodes {
			nodeMap[node.Hostname] = node
		}
	}

	return nodeMap, nil
}

func GetControlNodes() ([]Node, error) {
	controllers := []Node{}

	nodes, err := GetNodesByRole("control")
	if err == nil && len(nodes) > 0 {
		controllers = append(controllers, nodes...)
	}

	nodes, err = GetNodesByRole("control-converged")
	if err == nil && len(nodes) > 0 {
		controllers = append(controllers, nodes...)
	}

	if len(controllers) == 0 {
		return nil, fmt.Errorf(
			"failed to get control nodes(control or control-converged): %s",
			err.Error(),
		)
	}

	return controllers, nil
}

func GetPeerControlNodes() ([]Node, error) {
	controllers := []Node{}

	nodes, err := GetNodesByRole("control")
	if err == nil && len(nodes) > 0 {
		controllers = append(controllers, nodes...)
	}

	nodes, err = GetNodesByRole("control-converged")
	if err == nil && len(nodes) > 0 {
		controllers = append(controllers, nodes...)
	}

	if len(controllers) == 0 {
		return nil, fmt.Errorf(
			"failed to get control nodes(control or control-converged): %s",
			err.Error(),
		)
	}

	for i, controller := range controllers {
		if controller.IsLocal() {
			controllers = slices.Delete(controllers, i, i+1)
			break
		}
	}

	return controllers, nil
}

func GetOneOfControllerNode() (*Node, error) {
	nodes, err := GetControlNodes()
	if err != nil {
		return nil, err
	}

	return &nodes[0], nil
}

func GetNodeByHostname(hostname string) (*Node, error) {
	for _, node := range ListNodes() {
		if node.Hostname == hostname {
			return &node, nil
		}
	}

	return nil, fmt.Errorf(
		"failed to get node by hostname %s",
		hostname,
	)
}

func GetNodesFromRoles() []Node {
	roleNodes := []Node{}
	for _, role := range Roles {
		role := GetRole(role)
		if role == nil {
			continue
		}

		roleNodes = append(roleNodes, role.Nodes...)
	}

	return roleNodes
}

func GetActiveNodeMap() map[string]Node {
	nodeMap := map[string]Node{}
	for _, node := range GetNodesFromRoles() {
		nodeMap[node.Hostname] = node
	}

	return nodeMap
}

func SetNodeDetails(nodesWithDetails []Node) {
	UpdateNodes.Lock()
	defer UpdateNodes.Unlock()
	nodes = nodesWithDetails
}

func ListNodes() []Node {
	return nodes
}
