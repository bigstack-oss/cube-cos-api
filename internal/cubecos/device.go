package cubecos

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func AddDevice(req nodes.DeviceReqOpts) error {
	out, err := exec.Command("hex_sdk", "ceph_osd_add_disk_raw", req.Device).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to execute device adding cmd %s(%v %s)",
			req.Device, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to add device %s(%v %s)",
			req.Device, err, string(out),
		)
	}

	return WaitDeviceToBeAdded(req, 180)
}

func PromoteOrDemoteDevice(req nodes.DeviceReqOpts) error {
	promoteOrDemote := "ceph_osd_promote_disk"
	if strings.EqualFold(req.Class, blockdevice.HDD) {
		promoteOrDemote = "ceph_osd_demote_disk"
	}

	out, err := exec.Command("hex_sdk", promoteOrDemote, req.Device).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to execute device update cmd %s(%v %s)",
			req.Device, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to update device %s(%v %s)",
			req.Device, err, string(out),
		)
	}

	return nil
}

func RemoveDevice(req nodes.DeviceReqOpts) error {
	out, err := exec.Command("hex_sdk", "ceph_osd_remove_disk", req.Device, "force").CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to remove device cmd %s(%v %s)",
			req.Device, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to remove device %s(%v %s)",
			req.Device, err, string(out),
		)
	}

	return nil
}

func GetOsdsByHostDevice(req nodes.DeviceReqOpts) ([]ceph.Osd, error) {
	deviceMap, err := ceph.GetDeviceMapByHost(req.Hostname)
	if err != nil {
		return nil, err
	}

	dev := blockdevice.WithDevPath(req.Device)
	device, found := deviceMap[dev]
	if !found {
		return nil, fmt.Errorf(
			"device %s not found on host %s",
			req.Device, req.Hostname,
		)
	}
	if len(device.Osds) == 0 {
		return nil, fmt.Errorf(
			"device %s has no OSDs on host %s",
			req.Device, req.Hostname,
		)
	}

	return device.Osds, nil
}

func WaitDeviceToBeAdded(req nodes.DeviceReqOpts, timeout int) error {
	for range timeout {
		wait.Seconds(1)

		osds, err := GetOsdsByHostDevice(req)
		if err != nil {
			continue
		}

		if len(osds) <= 0 {
			continue
		}

		err = WaitOsdsStatus(osds, status.Up, 2)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf(
		"failed to wait device %s to be added on the %s after 120 seconds",
		req.Device,
		req.Hostname,
	)
}
