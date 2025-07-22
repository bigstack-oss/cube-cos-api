package cubecos

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func WaitOsdStatus(id, status string, timeout int) error {
	for range timeout {
		wait.Seconds(1)

		out, err := exec.Command("ceph", "osd", "df", id, "-f", "json").CombinedOutput()
		if err != nil {
			err := fmt.Errorf("failed to execute osd df for status(%v %s)", err, string(out))
			log.Errorf("nodes: %v", err)
			continue
		}

		if !IsHexSdkSuccess(err) {
			err := fmt.Errorf("failed to wait %s status(%v %s)", id, err, string(out))
			log.Errorf("nodes: %v", err)
			continue
		}

		raw := &ceph.RawOsd{}
		err = json.Unmarshal(out, raw)
		if err != nil {
			log.Errorf("ceph: failed to unmarshal osd df output %s(%v)", string(out), err)
			continue
		}
		if len(raw.Nodes) == 0 {
			log.Errorf("ceph: no nodes found for osd %s", id)
			continue
		}

		if strings.EqualFold(raw.Nodes[0].Status, status) {
			log.Infof("ceph: %s is in status %s successfully", id, status)
			return nil
		}
	}

	return fmt.Errorf(
		"failed to wait %s status %s after %d seconds",
		id, status, timeout,
	)
}

func RestartOsd(req nodes.OsdReqOpts) error {
	service := fmt.Sprintf("ceph-osd@%s", strings.TrimPrefix(req.OsdId, "osd."))
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
	out, err := exec.Command("hex_sdk", "ceph_osd_remove", req.OsdId, "force").CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to remove osd cmd %s(%v %s)",
			req.OsdId, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to remove osd %s(%v %s)",
			req.OsdId, err, string(out),
		)
	}

	return nil
}

func ReweightOsd(req nodes.OsdReqOpts) error {
	value := fmt.Sprintf("%f", req.Reweight)
	out, err := exec.Command("ceph", "osd", "crush", "reweight", req.OsdId, value).CombinedOutput()
	if err != nil {
		log.Errorf(
			"hexSdk: failed to execute osd reweight cmd %s(%v %s)",
			req.OsdId, err, string(out),
		)
		return err
	}

	if !IsHexSdkSuccess(err) {
		return fmt.Errorf(
			"failed to reweight osd %s(%v %s)",
			req.OsdId, err, string(out),
		)
	}

	return nil
}
