package nodes

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go-micro.dev/v5/registry"
)

const (
	Module                         = "nodes"
	Db                             = "nodes"
	ReqCollection                  = "requests"
	CollectionTemporaryNodeDetails = "temporaryNodeDetails"

	CollectionIpmiAccess  = "ipmiAccess"
	CollectionIpmiSupport = "ipmiSupport"

	ReqDeviceCollection = "deviceRequests"
)

var (
	list         = atomic.Pointer[[]Node]{}
	previousList = atomic.Pointer[[]Node]{}

	serviceList = sync.Mutex{}
	nodeList    = sync.Mutex{}
)

type Node struct {
	Id           string `json:"id" yaml:"id" bson:"id"`
	SerialNumber string `json:"serialNumber" yaml:"serialNumber" bson:"serialNumber"`
	BoardSerial  string `json:"boardSerial" yaml:"boardSerial" bson:"boardSerial"`
	DataCenter   string `json:"dataCenter" yaml:"dataCenter" bson:"dataCenter"`
	Hostname     string `json:"hostname" yaml:"hostname" bson:"hostname"`
	Role         string `json:"role" yaml:"role" bson:"role"`

	Protocol     string `json:"protocol,omitempty" yaml:"protocol,omitempty" bson:"protocol,omitempty"`
	Address      string `json:"address" yaml:"address" bson:"address"`
	Ip           string `json:"ip" yaml:"ip" bson:"ip"`
	ManagementIP string `json:"managementIP" yaml:"managementIP" bson:"managementIP"`
	StorageIP    string `json:"storageIP" yaml:"storageIP" bson:"storageIP"`

	CpuSpec           string             `json:"cpuSpec" yaml:"cpuSpec" bson:"cpuSpec"`
	NetworkInterfaces []NetworkInterface `json:"networkInterfaces" yaml:"networkInterfaces" bson:"networkInterfaces"`
	BlockDevices      []BlockDevice      `json:"blockDevices" yaml:"blockDevices" bson:"blockDevices"`
	Vcpu              metric.Compute     `json:"vcpu" yaml:"vcpu" bson:"vcpu"`
	Memory            metric.Space       `json:"memory" yaml:"memory" bson:"memory"`
	Storage           metric.Space       `json:"storage" yaml:"storage" bson:"storage"`

	IsVirtualIpOwner bool `json:"isVirtualIpOwner" yaml:"isVirtualIpOwner" bson:"isVirtualIpOwner"`
	IpmiEnablement   `json:"ipmi" yaml:"ipmi" bson:"ipmi"`

	License       licenses.License `json:"license" yaml:"license,omitempty" bson:"license,omitempty"`
	Status        string           `json:"status" yaml:"status" bson:"status" default:"down"`
	UptimeSeconds float64          `json:"uptimeSeconds" yaml:"uptimeSeconds" bson:"uptimeSeconds"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" bson:"labels,omitempty"`
}

func (n *Node) IsLocal() bool {
	return n.Address == base.AdvertiseAddr && n.Hostname == base.Hostname
}

func (n *Node) IsUp() bool {
	return n.Status == status.Up
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

// note:
// in the current search lib(bleve), we realize that it seems like not able to serach the status keyword
// from the single status field, but it's pretty tricky that it will work if we add the status behind any of fields below
// like hostname, address, managementIP... so that's why we do this for the time being
// in the M2, we will try to deep dive into the bleve and see if we can find a better way to do this
func (n *Node) GenSearchableObject() Node {
	return Node{
		Hostname:     search.NormalizedKeyword(n.Hostname) + search.NormalizedKeyword(n.Status),
		Role:         search.NormalizedKeyword(n.Role),
		Address:      search.NormalizedKeyword(n.Address),
		Ip:           search.NormalizedKeyword(n.Ip),
		StorageIP:    search.NormalizedKeyword(n.StorageIP),
		ManagementIP: search.NormalizedKeyword(n.ManagementIP),
		License: licenses.License{
			Expiry: licenses.Expiry{
				Date: search.NormalizedKeyword(n.License.Expiry.Date),
			},
		},
	}
}

func IsLocal(hostname string) bool {
	return base.Hostname == hostname
}

func IsLocalAddress(address string) bool {
	return base.AdvertiseAddr == address
}

func New(node *registry.Node) Node {
	return Node{
		Role:         node.Metadata["role"],
		Id:           node.Id,
		SerialNumber: node.Metadata["serialNumber"],
		BoardSerial:  node.Metadata["boardSerial"],
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

func GetDiscoveredServices() ([]*registry.Service, error) {
	serviceList.Lock()
	defer serviceList.Unlock()

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

func GetControlNodes() ([]Node, error) {
	controllers := []Node{}

	nodes, err := GetNodesByRole(RoleControl)
	if err == nil && len(nodes) > 0 {
		controllers = append(controllers, nodes...)
	}

	nodes, err = GetNodesByRole(RoleControlConverged)
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
	controllers, err := GetControlNodes()
	if err != nil {
		return nil, err
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
	prev := list.Swap(&nodes)
	if prev != nil {
		previousList.Store(prev)
	}
}

func Sync() {
	SyncEachRole()
	SyncRoleCombination()
}

func ListPrevious() []Node {
	nodes := previousList.Load()
	if nodes == nil {
		return []Node{}
	}

	return *nodes
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

func Lock() {
	nodeList.Lock()
}

func Unlock() {
	nodeList.Unlock()
}
