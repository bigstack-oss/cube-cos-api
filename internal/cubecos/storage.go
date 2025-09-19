package cubecos

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/metric"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"github.com/dustin/go-humanize"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func GenCreateOptsByReqOpts(reqOpts images.ReqOpts) (*images.CreateOpts, error) {
	poolType := "glance-images"
	visibility := reqOpts.Visibility
	if reqOpts.SourceFromAnotherHypervisor {
		poolType = "cinder-volumes"
		visibility = "public"
	}

	storageBackend, err := GetStorageBackendByPoolType(poolType)
	if err != nil {
		return nil, err
	}

	return &images.CreateOpts{
		Dir:            images.GlanceDir,
		File:           reqOpts.File,
		Name:           reqOpts.Name,
		AttributesType: "default",
		Destination:    reqOpts.Destination,
		Domain:         reqOpts.Domain,
		PoolType:       poolType,
		StorageBackend: fmt.Sprintf(`"%s"`, storageBackend),
		Project:        reqOpts.Project,
		Visibility:     visibility,
		StreamingLogs:  make(chan float64),
		ReservedType:   reqOpts.Reserved.Type,
	}, nil
}

func IsDefaultStorage(name string) bool {
	file, err := os.Open(storages.CinderConf)
	if err != nil {
		log.Errorf("storages: failed to read cinder conf file(%v)", err)
		return false
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	defaultStorage := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "default_volume_type") {
			continue
		}

		segments := strings.SplitN(line, "=", 2)
		if len(segments) != 2 {
			continue
		}

		defaultStorage = strings.TrimSpace(segments[1])
	}

	err = scanner.Err()
	if err != nil {
		log.Errorf("storages: error reading cinder conf file(%v)", err)
		return false
	}

	return defaultStorage == name
}

func GetStorageBackendByPoolType(poolType string) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxMinutes(3))
	defer cancel()
	out, err := exec.CommandContext(ctx, "hex_sdk", "os_list_volume_backend_by_pool", poolType).CombinedOutput()
	if err != nil {
		err := genIntegrationErr("storage backend exec failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return "", err
	}

	if !IsHexSuccessful(err) {
		err := genIntegrationErr("storage backend output failure")
		log.Errorf("storage: %s (%s)", err.Error(), string(out))
		return "", err
	}

	storageBackend := strings.TrimSpace(string(out))
	if len(storageBackend) == 0 {
		err := genIntegrationErr("storage backend empty output")
		log.Errorf("storage: %s", err.Error())
		return "", err
	}

	return storageBackend, nil
}

func GetCephUsage() metric.Space {
	b, err := exec.Command("ceph", "df", "-f", "json").Output()
	if err != nil {
		log.Errorf("metrics: failed to get ceph usage(%v)", err)
		return metric.Space{}
	}

	cephUsage := ceph.SpaceMetrics{}
	err = json.Unmarshal(b, &cephUsage)
	if err != nil {
		log.Errorf("metrics: failed to unmarshal ceph usage(%v)", err)
		return metric.Space{}
	}

	total := float64(cephUsage.Stats.TotalBytes) / 1024.0 / 1024.0
	used := float64(cephUsage.Stats.TotalUsedBytes) / 1024.0 / 1024.0
	avail := float64(cephUsage.Stats.TotalAvailBytes) / 1024.0 / 1024.0
	return metric.Space{
		TotalMiB:    math.RoundDown(total, 4),
		UsedMiB:     math.RoundDown(used, 4),
		FreeMiB:     math.RoundDown(avail, 4),
		UsedPercent: math.RoundDown((used/total)*100.0, 4),
		FreePercent: math.RoundDown((avail/total)*100.0, 4),
	}
}

func GetRawBlockDevices() ([]nodes.RawBlockDevice, error) {
	out, err := exec.Command("/bin/lsblk", "--sort", "name", "--json", "-o", "TYPE,NAME,ROTA,SERIAL,SIZE,MOUNTPOINTS", "-e", blockdevice.NetCode).Output()
	if err != nil {
		log.Errorf("nodes: failed to get block device info(%v)", err)
		return nil, err
	}

	resp := map[string][]nodes.RawBlockDevice{}
	err = json.Unmarshal(out, &resp)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal block device info(%v)", err)
		return nil, err
	}

	raws, found := resp["blockdevices"]
	if !found {
		log.Errorf("nodes: failed to find block devices in the output")
		return nil, err
	}
	if len(raws) <= 0 {
		log.Errorf("nodes: no block device found")
		return nil, errors.New("no block device found")
	}

	return getBlockOrPartitionOnly(raws), nil
}

func RawToBlockDevice(rawBlockDev nodes.RawBlockDevice) nodes.BlockDevice {
	return nodes.BlockDevice{
		Serial:  rawBlockDev.Serial,
		Name:    rawBlockDev.Name,
		Type:    convertBlockDeviceType(rawBlockDev.Rota),
		SizeMiB: convertBlockDeviceSize(rawBlockDev.Size),
		Status:  status.BlockDevice{Current: "can be added"},
	}
}

func getBlockOrPartitionOnly(rawBlockDevs []nodes.RawBlockDevice) []nodes.RawBlockDevice {
	blockDevs := []nodes.RawBlockDevice{}
	for _, rawBlockDev := range rawBlockDevs {
		if rawBlockDev.IsBlock() {
			blockDevs = append(blockDevs, rawBlockDev)
		}

		if rawBlockDev.IsPartition() {
			blockDevs = append(blockDevs, rawBlockDev)
		}
	}

	return blockDevs
}

func convertBlockDeviceType(rotation bool) string {
	if rotation {
		return blockdevice.HDD
	}

	return blockdevice.SSD
}

func convertBlockDeviceSize(sizeStr string) float64 {
	bytes, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		log.Errorf("nodes: failed to convert block device size(%v)", err)
		return 0
	}

	sizeMiB := float64(bytes) / (1024.0 * 1024.0)
	return math.RoundDown(sizeMiB, 4)
}
