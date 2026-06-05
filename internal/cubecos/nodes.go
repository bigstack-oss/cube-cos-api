package cubecos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v1/accelerators/devices"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/gpu"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
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

func IsVirtualIpOwner(hostname string) bool {
	node, err := GetVirtualIpController()
	if err != nil {
		return false
	}

	return node.Hostname == hostname
}

func GetVirtualIpController() (*nodes.Node, error) {
	nodes := nodes.List()
	if len(nodes) == 0 {
		return nil, errors.New(
			"no nodes found in the system",
		)
	}

	if !base.IsHaEnabled {
		return &nodes[0], nil
	}

	syncVirutalIpOwner(&nodes)
	if !base.IsHaEnabled {
		return &nodes[0], nil
	}

	for _, node := range nodes {
		if node.IsVirtualIpOwner {
			return &node, nil
		}
	}

	return nil, errors.New(
		"failed to get virtual IP controller, no node is virtual IP owner",
	)
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

func GetPrimaryControllerHost() (string, error) {
	hostsStr, err := GetTuningValue(CubeSysControllerHosts)
	if err != nil {
		return "", err
	}

	if hostsStr == "" {
		return "", fmt.Errorf("controller hosts is empty")
	}

	hosts := strings.Split(hostsStr, ",")
	if len(hosts) == 0 {
		return "", fmt.Errorf("no controller hosts found")
	}

	return hosts[0], nil
}

func IsPrimaryController(hostname string) bool {
	primary, err := GetPrimaryControllerHost()
	if err != nil {
		log.Errorf("nodes: failed to get primary controller host(%v)", err)
		return false
	}

	return primary == hostname
}

func ListNodesWithTimeSensitiveInfo() []nodes.Node {
	list := nodes.List()
	if len(list) == 0 {
		return []nodes.Node{}
	}

	syncTimeSensitiveInfo(&list)
	backfillMissingInfo(&list)
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

func DrainNode() error {
	SyncFirmwareUpgradeProgressToAllNodes()
	if !IsVirtualIpOwner(base.Hostname) {
		return nil
	}

	err := MoveVirtualIpOwner()
	if err != nil {
		log.Errorf("nodes: failed to move virtual ip owner(%v)", err)
		return err
	}

	err = WaitForVirutalIpOwnerChanged(base.Hostname)
	if err != nil {
		log.Errorf("nodes: failed to wait for virtual ip owner changed(%v)", err)
		return err
	}

	return nil
}

func WaitForVirutalIpOwnerChanged(oldOwner string) error {
	for range 600 {
		wait.Seconds(1)
		host, err := pacemaker.GetVirtualIpHost()
		if err != nil {
			log.Errorf("nodes: failed to get virtual ip host(%v)", err)
			continue
		}

		if host == oldOwner {
			log.Infof("nodes: virtual ip owner is still %s, wait for it changed", oldOwner)
			continue
		}

		return nil
	}

	return fmt.Errorf(
		"failed to wait for virtual ip owner changed in 10 minutes",
	)
}

func syncTimeSensitiveInfo(list *[]nodes.Node) {
	syncLicense(list)
	syncVirutalIpOwner(list)
	syncPowerStatus(list)
	syncIpmiAccess(list)
}

func backfillMissingInfo(list *[]nodes.Node) {
	backfillMissingInfraSpec(list)
	backfillMissingStatus(list)
}

func backfillMissingInfraSpec(list *[]nodes.Node) {
	for i, node := range *list {
		if node.NetworkInterfaces == nil {
			(*list)[i].NetworkInterfaces = []nodes.NetworkInterface{}
		}

		if node.BlockDevices == nil {
			(*list)[i].BlockDevices = []nodes.BlockDevice{}
		}
	}
}

func backfillMissingStatus(list *[]nodes.Node) {
	for i, node := range *list {
		if node.Status == "" {
			(*list)[i].Status = status.Syncing
		}
	}
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
		if !IsNodeHasPowerRequest(node.Hostname) {
			continue
		}

		status, err := getPendingPowerStatus(node.Hostname)
		if err != nil {
			log.Errorf("nodes: failed to get node(%s) pending power status(%v)", node.Hostname, err)
			continue
		}

		(*list)[i].Status = status
	}
}

func syncIpmiAccess(list *[]nodes.Node) {
	for i, node := range *list {
		(*list)[i].IpmiEnablement.IsSupported = hasIpmiSupportRecord(node.Hostname)
		(*list)[i].IpmiEnablement.IsConnected = hasIpmiRecord(node.Hostname)
	}
}

func IsNodeHasPowerRequest(hostname string) bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		nodes.Db,
		nodes.ReqCollection,
		bson.M{"hostname": hostname},
	)
	if err != nil {
		log.Errorf("nodes: failed to get node(%s) power request count(%v)", hostname, err)
		return false
	}

	return count > 0
}

func hasIpmiSupportRecord(hostname string) bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		nodes.Db,
		nodes.CollectionIpmiSupport,
		bson.M{"host": hostname, "supported": true},
	)
	if err != nil {
		log.Errorf("nodes: failed to get node(%s) ipmi support record(%v)", hostname, err)
		return false
	}

	return count > 0
}

func hasIpmiRecord(hostname string) bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		nodes.Db,
		nodes.CollectionIpmiAccess,
		bson.M{"host": hostname},
	)
	if err != nil {
		log.Errorf("nodes: failed to get node(%s) ipmi record(%v)", hostname, err)
		return false
	}

	return count > 0
}

func getPendingPowerStatus(hostname string) (string, error) {
	mongo := mongo.GetGlobalHelper()
	doc, err := mongo.Get(
		nodes.Db,
		nodes.ReqCollection,
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

// Returns a map of GPUs from hex, with PCI address as key.
func GetNodeGpusMap(nodeName string) (map[string]gpu.GpuFromHex, error) {
	gpuMap := map[string]gpu.GpuFromHex{}
	gpus, err := listNodeGpus(nodeName)

	if err != nil {
		return nil, err
	}

	for _, gpu := range gpus {
		gpuMap[gpu.PciAddress] = gpu
	}

	return gpuMap, nil
}

func listNodeGpus(nodeName string) ([]gpu.GpuFromHex, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	out, err := exec.CommandContext(ctx, "hex_sdk", "gpu_device_list").CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to list gpus for node %s via hex_sdk: %v", nodeName, err)
		return nil, err
	}

	if !IsHexSuccessful(err) {
		log.Errorf("nodes: output error when listing gpus for node %s via hex_sdk: %v", nodeName, err)
		return nil, err
	}

	gpus := []gpu.GpuFromHex{}
	err = json.Unmarshal(out, &gpus)
	if err != nil {
		log.Errorf("nodes: failed to parse output when listing gpus for node %s via hex_sdk: %v", nodeName, err)
		return nil, err
	}

	return gpus, nil
}

func GetNodeVgpuProfilesMap(gpuId string) map[uint32]gpu.VgpuProfileFromHex {
	profiles := listNodeVgpuProfiles(gpuId)
	profilesMap := map[uint32]gpu.VgpuProfileFromHex{}

	for _, profile := range *profiles {
		profilesMap[profile.Id] = profile
	}

	return profilesMap
}

func listNodeVgpuProfiles(gpuId string) *[]gpu.VgpuProfileFromHex {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	profiles := []gpu.VgpuProfileFromHex{}

	out, err := exec.CommandContext(ctx, "hex_sdk", "gpu_vgpu_profile_list", "-gpuId", gpuId).CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to list vgpu profiles for gpu %s: %v", gpuId, err)
		return &profiles
	}

	if !IsHexSuccessful(err) {
		log.Errorf("nodes: output error when listing vgpu profiles for gpu %s via hex_sdk: %v", gpuId, err)
		return &profiles
	}

	err = json.Unmarshal(out, &profiles)
	if err != nil {
		log.Errorf("nodes: failed to parse output when listing vgpu profiles for gpu %s via hex_sdk: %v", gpuId, err)
		return &profiles
	}

	return &profiles
}

func UpdateNodeGpuCard(gpuId string, req gpu.UpdateGpuCardRequest) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	args := []string{"gpu_device_update", "-gpuId", gpuId, "-resourceType", string(req.ResourceType)}
	if len(req.Profiles) > 0 {
		profilesJson, err := json.Marshal(req.Profiles)
		if err != nil {
			return fmt.Errorf("failed to marshal profiles: %v", err)
		}
		args = append(args, "-profiles", string(profilesJson))
	}

	out, err := exec.CommandContext(ctx, "hex_sdk", args...).CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to update gpu card %s via hex_sdk: %v, output: %s", gpuId, err, string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		log.Errorf("nodes: output error when updating gpu card %s via hex_sdk: %s", gpuId, string(out))
		return fmt.Errorf("hex_sdk gpu_device_update failed: %s", string(out))
	}

	return nil
}
