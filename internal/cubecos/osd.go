package cubecos

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func RestartOsd(req nodes.OsdReqOpts) error {
	service := fmt.Sprintf("ceph-osd@%s", strings.TrimPrefix(req.Id, "osd."))
	out, err := exec.Command("systemctl", "restart", service).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to execute osd restart cmd %s(%v %s)",
			service, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to restart osd %s(%v %s)",
			service, err, string(out),
		)
	}

	return nil
}

func RemoveOsd(req nodes.OsdReqOpts) error {
	out, err := exec.Command("hex_sdk", "ceph_osd_remove", req.Id).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to remove osd cmd %s(%v %s)",
			req.Id, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to remove osd %s(%v %s)",
			req.Id, err, string(out),
		)
	}

	return nil
}

func ReweightOsd(req nodes.OsdReqOpts) error {
	id := fmt.Sprintf("osd.%s", req.Id)
	reweight := fmt.Sprintf("%f", req.Reweight)
	out, err := exec.Command("ceph", "osd", "crush", "reweight", id, reweight).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to execute osd reweight cmd %s(%v %s)",
			req.Id, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to reweight osd %s(%v %s)",
			req.Id, err, string(out),
		)
	}

	return nil
}
