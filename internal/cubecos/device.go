package cubecos

import (
	"fmt"
	"os/exec"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func AddDevice(req nodes.DeviceReqOpts) error {
	out, err := exec.Command("hex_sdk", "ceph_osd_add_disk_raw", req.Device).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to add device cmd %s(%v %s)",
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
