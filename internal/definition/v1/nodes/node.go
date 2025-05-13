package nodes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

const (
	Module = "nodes"
)

var (
	list           = atomic.Pointer[[]Node]{}
	updateServices = sync.Mutex{}
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
	License           licenses.License   `json:"license" yaml:"license,omitempty" bson:"license,omitempty"`
	Status            string             `json:"status" yaml:"status"`
	CpuSpec           string             `json:"cpuSpec" yaml:"cpuSpec" bson:"cpuSpec"`
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces" yaml:"networkInterfaces" bson:"networkInterfaces"`
	BlockDevices      []BlockDevice      `json:"blockDevices" yaml:"blockDevices" bson:"blockDevices"`
	Vcpu              metric.Compute     `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
	Memory            metric.Space       `json:"memory" yaml:"memory" bson:"memory"`
	Storage           metric.Space       `json:"storage" yaml:"storage" bson:"storage"`
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

// note:
// rota is named by lsblk tool, it means rotational device like HDD
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

func (n *Node) GetNodeUrl() string {
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

func (n *Node) PostLicenseUrl() string {
	u := url.URL{
		Scheme: n.Protocol,
		Host:   n.Address,
		Path:   fmt.Sprintf("/api/v1/datacenters/%s/licenses/hosts/%s", base.DataCenterName, n.Hostname),
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

func (n *Node) PatchTuningUrl(tuning string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/parameters/%s", base.DataCenterName, tuning)
	return u.String()
}

func (n *Node) PatchTuningTaskUrl(id string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/tunings/tasks/%s", base.DataCenterName, id)
	return u.String()
}

func (n *Node) PatchTriggerTaskUrl(trigger triggers.ApiSchema) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/triggers/tasks/%s", base.DataCenterName, trigger.Name)
	return u.String()
}

func (n *Node) CreateSupportFileUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles", base.DataCenterName)
	return u.String()
}

func (n *Node) PatchSupportFileTaskUrl(file support.File) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/supportFiles/%s", base.DataCenterName, file.Group)
	return u.String()
}

func (n *Node) PatchSettingTaskUrl(setting settings.Setting) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/settings/tasks", base.DataCenterName)
	return u.String()
}

func (n *Node) DeleteRepairingTaskUrl() string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/healths/tasks/repairing", base.DataCenterName)
	return u.String()
}

func (n *Node) DeleteModuleRepairingTaskUrl(module string) string {
	u := url.URL{}
	u.Scheme = n.Protocol
	u.Host = n.Address
	u.Path = fmt.Sprintf("/api/v1/datacenters/%s/healths/tasks/repairing/%s", base.DataCenterName, module)
	return u.String()
}

func (n *Node) IsLocal() bool {
	return n.Address == base.AdvertiseAddr && n.Hostname == base.Hostname
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

func IsLocal(hostname string) bool {
	return base.Hostname == hostname
}

func IsLocalAddress(address string) bool {
	return base.AdvertiseAddr == address
}

func GenerateNodeHashByMacAddr() (string, error) {
	macAddr, err := base.GetMacAddr(base.NetMajorInterface)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(macAddr))
	return hex.EncodeToString(hash[:])[:8], nil
}

func GetNodesByRole(role string) ([]Node, error) {
	svcs, err := GetDiscoveredServices()
	if err != nil {
		return nil, err
	}

	list := []Node{}
	for _, svc := range svcs {
		nodes := parseNodesByRole(svc, role)
		if len(nodes) != 0 {
			list = append(list, nodes...)
		}
	}

	return list, nil
}

func parseNodesByRole(svc *registry.Service, role string) []Node {
	nodes := []Node{}
	for _, node := range svc.Nodes {
		if node.Metadata["role"] != role {
			continue
		}

		nodes = append(nodes, New(node))
	}

	return nodes
}

func GetDiscoveredServices() ([]*registry.Service, error) {
	updateServices.Lock()
	defer updateServices.Unlock()

	svcs, err := registry.GetService(base.ServiceDiscoveryIdentity)
	if err != nil {
		log.Errorf("nodes: failed to get service from %s(%v)", base.ServiceDiscoveryIdentity, err)
		return nil, err
	}

	if len(svcs) <= 0 {
		err := fmt.Errorf("no any service find from %s", base.ServiceDiscoveryIdentity)
		log.Errorf(err.Error())
		return nil, err
	}

	return svcs, nil
}

func HostnameMap() (map[string]Node, error) {
	svcs, err := GetDiscoveredServices()
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

func parseNodes(svc *registry.Service) []Node {
	nodes := []Node{}

	for _, node := range svc.Nodes {
		if IsLocal(node.Metadata["hostname"]) {
			continue
		}

		nodes = append(nodes, New(node))
	}

	return nodes
}

func New(node *registry.Node) Node {
	return Node{
		Role:         node.Metadata["role"],
		Id:           node.Id,
		SerialNumber: node.Metadata["serialNumber"],
		DataCenter:   node.Metadata["dataCenter"],
		Protocol:     node.Metadata["protocol"],
		Ip:           node.Metadata["ip"],
		Hostname:     node.Metadata["hostname"],
		Address:      node.Address,
		Labels: map[string]string{
			"isGpuEnabled": node.Metadata["isGpuEnabled"],
		},
	}
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
			"failed to get control nodes(control or control-converged): %v",
			err,
		)
	}

	return controllers, nil
}

func GetPeerControls() ([]Node, error) {
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
			"failed to get control nodes(control or control-converged): %v",
			err,
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

func GetController() (*Node, error) {
	nodes, err := GetControlNodes()
	if err != nil {
		return nil, err
	}

	return &nodes[0], nil
}

func List() []Node {
	nodes := list.Load()
	if nodes == nil {
		return []Node{}
	}

	return *nodes
}

func Get(hostname string) (*Node, error) {
	for _, node := range List() {
		if node.Hostname == hostname {
			return &node, nil
		}
	}

	return nil, fmt.Errorf(
		"failed to get node by hostname %s",
		hostname,
	)
}

func SetList(nodes []Node) {
	list.Swap(&nodes)
}

func Sync() {
	SyncEachRole()
	SyncRoleCombination()
}

func SyncEachRole() {
	for _, role := range roles {
		nodes, err := GetNodesByRole(role)
		if err != nil {
			return
		}

		role := Role{Name: role, Nodes: nodes, Hosts: convertToHosts(nodes)}
		switch role.Name {
		case RoleControl:
			Control.Swap(&role)
		case RoleCompute:
			Compute.Swap(&role)
		case RoleStorage:
			Storage.Swap(&role)
		case RoleControlConverged:
			ControlConverged.Swap(&role)
		case RoleModerator:
			Moderator.Swap(&role)
		case RoleEdgeCore:
			EdgeCore.Swap(&role)
		}
	}
}

func SyncRoleCombination() {
	SetAllRoles()
	SetAllGeneralRoles()
	SetControlRoles()
	SetComputeRoles()
}

func GetMap() map[string]Node {
	nodes := map[string]Node{}
	for _, node := range GetNodesByRoles() {
		nodes[node.Hostname] = node
	}

	return nodes
}

func GetNodesByRoles() []Node {
	nodes := []Node{}
	for _, role := range roles {
		switch role {
		case RoleControl:
			nodes = append(nodes, Control.Load().Nodes...)
		case RoleCompute:
			nodes = append(nodes, Compute.Load().Nodes...)
		case RoleStorage:
			nodes = append(nodes, Storage.Load().Nodes...)
		case RoleControlConverged:
			nodes = append(nodes, ControlConverged.Load().Nodes...)
		case RoleModerator:
			nodes = append(nodes, Moderator.Load().Nodes...)
		case RoleEdgeCore:
			nodes = append(nodes, EdgeCore.Load().Nodes...)
		}
	}

	return nodes
}

func convertToHosts(nodes []Node) []Host {
	hosts := []Host{}
	for _, node := range nodes {
		hosts = append(
			hosts,
			Host{
				Name: node.Hostname,
				Ip:   node.Ip,
			},
		)
	}

	return hosts
}
