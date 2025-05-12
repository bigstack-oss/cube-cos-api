package cubecos

import (
	"errors"
	"os/exec"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/math"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/dustin/go-humanize"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func GetRawBlockDevices() ([]nodes.RawBlockDevice, error) {
	b, err := exec.Command("/bin/lsblk", "--sort", "name", "--json", "-o", "TYPE,NAME,ROTA,SERIAL,SIZE,MOUNTPOINTS", "-e", blockdevice.NetCode).Output()
	if err != nil {
		log.Errorf("nodes: failed to get block device info: %s", err.Error())
		return nil, err
	}

	blockDevMap := map[string][]nodes.RawBlockDevice{}
	err = json.Unmarshal(b, &blockDevMap)
	if err != nil {
		log.Errorf("nodes: failed to unmarshal block device info: %s", err.Error())
		return nil, err
	}

	rawBlockDevs, found := blockDevMap["blockdevices"]
	if !found {
		log.Errorf("nodes: failed to find block devices in the output")
		return nil, err
	}
	if len(rawBlockDevs) <= 0 {
		log.Errorf("nodes: no block device found")
		return nil, errors.New("no block device found")
	}

	return getBlockOrPartitionOnly(rawBlockDevs), nil
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

func ConvertToBlockDevice(rawBlockDev nodes.RawBlockDevice) nodes.BlockDevice {
	return nodes.BlockDevice{
		Serial:  rawBlockDev.Serial,
		Name:    rawBlockDev.Name,
		Type:    convertBlockDeviceType(rawBlockDev.Rota),
		SizeMiB: convertBlockDeviceSize(rawBlockDev.Size),
		Status:  status.BlockDevice{Current: "can be added"},
	}
}

func convertBlockDeviceType(rota bool) string {
	if rota {
		return "HDD"
	}

	return "SSD"
}

func convertBlockDeviceSize(sizeStr string) float64 {
	bytes, err := humanize.ParseBytes(sizeStr)
	if err != nil {
		log.Errorf("nodes: failed to convert block device size: %s", err.Error())
		return 0
	}

	sizeMiB := float64(bytes) / (1024.0 * 1024.0)
	return math.RoundDown(sizeMiB, 4)
}
