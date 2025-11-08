package firmwares

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operate(req *firmwares.ReqOpts) error {
	switch req.Status.Desired {
	case status.Installed:
		return cubecos.UpgradeFirmware(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for firmware %s",
		req.Status.Desired,
		req.Version,
	)
}

func (o *Operator) handleExit(req *firmwares.ReqOpts, err error) {
	if err != nil {
		log.Errorf("firmwares: failed to %s %s(%v)", req.Status.Desired, req.Version, err)
		req.SetError(err.Error())
	} else {
		log.Infof("firmwares: %s %s successfully", req.Status.Desired, req.Version)
		req.SetInstalled()
	}

	req.Hostname = base.Hostname
	o.reportToController(req)
}

func (o *Operator) reportToController(req *firmwares.ReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("firmwares: failed to report %s result to control(%v)", req.Version, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateFirmwareTaskUrl())
	if err != nil {
		log.Errorf("firmwares: failed to send firmware(%s) task update to %s(%v)", req.Version, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"firmwares: has error response from %s firmware %s task update(%v)",
			node.Hostname,
			req.Version,
			string(resp.Body()),
		)
	}
}
