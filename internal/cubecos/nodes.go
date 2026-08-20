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

// GetNodeGpuById returns the hex GPU whose Id matches gpuId, or gpu.ErrGpuNotFound.
// Reuses GetNodeGpusMap (keyed by PCI address) and matches on the GPU UUID Id.
func GetNodeGpuById(nodeName, gpuId string) (gpu.GpuFromHex, error) {
	gpusMap, err := GetNodeGpusMap(nodeName)
	if err != nil {
		return gpu.GpuFromHex{}, err
	}

	for _, g := range gpusMap {
		if g.Id == gpuId {
			return g, nil
		}
	}

	return gpu.GpuFromHex{}, fmt.Errorf("gpu %s not found on node %s: %w", gpuId, nodeName, gpu.ErrGpuNotFound)
}

// UpdateNodeGpuCard sets a GPU's resource type via hex_config. Profiles, when
// present, are passed as a single JSON-string positional argument.
//
// The timeout is deliberately far longer than the 30s used by the read-only
// hex_sdk calls in this file, because this call *mutates hardware*. A resource
// switch releases the card from vfio-pci, carves it (sriov-manage -e brings up
// 48 VFs; a MIG-backed carve enables MIG mode and creates GPU instances),
// rewrites /etc/nova/nova.d/gpu.conf and then restarts the nova services.
// Measured on cn13 (RTX PRO 6000 Blackwell, 2026-08-20): pgpu -> sriovVgpu took
// 37.7s and pgpu -> migBackedVgpu 30.4s, so 30s cut the first one off mid-flight
// and left the second passing only by luck.
//
// What the old timeout actually cost is worse than a slow request: CommandContext
// SIGKILLs hex_config on expiry, i.e. it kills a process partway through
// repartitioning a GPU, with the hardware, /etc/cube/cos/gpu/config.json and
// nova's config able to disagree afterwards. The point of the longer budget is to
// let that sequence finish, not merely to return a 200.
//
// Note the caller may still not see the result: haproxy fronting this API is
// configured with `timeout server 1m` and nginx's /cos-api/ location takes the
// 60s proxy_read_timeout default, so a switch that runs past a minute is reported
// to the client as a gateway timeout regardless of the value here - the work
// still completes. Making the response honest needs the endpoint to go
// asynchronous (return 202 and let the caller poll Status.IsProcessing, which
// the ReqGpuCollection record already drives); that is a separate change.
func UpdateNodeGpuCard(gpuId string, req gpu.UpdateGpuCardRequest) error {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(180))
	defer cancel()

	args := []string{"gpu_resource_set", gpuId, string(req.ResourceType)}
	if len(req.Profiles) > 0 {
		profilesJson, err := json.Marshal(req.Profiles)
		if err != nil {
			return fmt.Errorf("failed to marshal gpu profiles: %w", err)
		}
		args = append(args, string(profilesJson))
	}

	out, err := exec.CommandContext(ctx, "hex_config", args...).CombinedOutput()
	if err != nil {
		log.Errorf("nodes: failed to set gpu %s resource via hex_config: %v, output: %s", gpuId, err, string(out))
		return err
	}

	if !IsHexSuccessful(err) {
		log.Errorf("nodes: output error when setting gpu %s resource via hex_config: %s", gpuId, string(out))
		return fmt.Errorf("hex_config gpu_resource_set failed: %s", string(out))
	}

	return nil
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

// GetNodeVgpuProfilesMap returns the vGPU profiles for a GPU as both a
// profile-id map and the raw collection. The error is non-nil when the hex
// profile fetch itself failed (command or parse error), which callers treat as
// degraded enrichment rather than an empty-but-valid profile set; the map and
// collection are still returned (empty) so callers can proceed.
func GetNodeVgpuProfilesMap(gpuId string) (map[uint32]gpu.VgpuProfileFromHex, gpu.VgpuProfileCollectionFromHex, error) {
	collection, err := getNodeVgpuProfileCollection(gpuId)

	// Profile IDs will not conflict between SR-IOV and MIG.
	profilesMap := map[uint32]gpu.VgpuProfileFromHex{}

	if collection.Sriov != nil {
		for _, profile := range *collection.Sriov {
			profilesMap[profile.Id] = profile
		}
	}

	if collection.MigBacked != nil {
		for _, profile := range *collection.MigBacked {
			profilesMap[profile.Id] = profile
		}
	}

	return profilesMap, collection, err
}

func getNodeVgpuProfileCollection(gpuId string) (gpu.VgpuProfileCollectionFromHex, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	collection := gpu.VgpuProfileCollectionFromHex{}

	// Output (stdout only) rather than CombinedOutput: gpu_vgpu_profile_list
	// prints diagnostics to stderr, and any such line would be prepended to the
	// JSON and break the unmarshal below. Same reasoning as
	// GetNodePgpuAttachedInstance. stderr is still surfaced on failure via
	// ExitError.
	out, err := exec.CommandContext(ctx, "hex_sdk", "gpu_vgpu_profile_list", gpuId).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			log.Errorf("nodes: failed to list vgpu profiles for gpu %s: %v: %s", gpuId, err, exitErr.Stderr)
		} else {
			log.Errorf("nodes: failed to list vgpu profiles for gpu %s: %v", gpuId, err)
		}

		return collection, err
	}

	err = json.Unmarshal(out, &collection)
	if err != nil {
		log.Errorf("nodes: failed to parse output when listing vgpu profiles for gpu %s via hex_sdk: %v", gpuId, err)
		return collection, err
	}

	return collection, nil
}

// Returns (nil, nil) when the GPU has no attached instance.
func GetNodePgpuAttachedInstance(pciAddress string) (*gpu.PgpuAttachedInstanceFromHex, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	attachedInstance := gpu.PgpuAttachedInstanceFromHex{}

	// Output (stdout only) rather than CombinedOutput: a warning line printed to
	// stderr ahead of the JSON would otherwise corrupt the unmarshal below.
	out, err := exec.CommandContext(ctx, "hex_sdk", "gpu_pgpu_attached_instance_get", pciAddress).Output()
	if err != nil {
		return nil, fmt.Errorf("nodes: failed to get attached instance for gpu %s via hex_sdk: %w", pciAddress, err)
	}

	err = json.Unmarshal(out, &attachedInstance)
	if err != nil {
		return nil, fmt.Errorf("nodes: failed to parse output when getting attached instance for gpu %s via hex_sdk: %w", pciAddress, err)
	}

	if len(attachedInstance.Id) == 0 {
		return nil, nil
	}

	return &attachedInstance, nil
}
