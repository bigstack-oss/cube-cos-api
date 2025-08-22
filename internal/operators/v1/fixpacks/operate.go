package fixpacks

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operate(req *fixpacks.ReqOpts) error {
	switch req.Status.Desired {
	case status.Installed:
		return cubecos.InstallFixpack(req)
	case status.Rollbacked:
		return cubecos.RollbackFixpack()
	}

	return fmt.Errorf(
		"unknown desired action(%s) for fixpack %s",
		req.Status.Desired,
		req.Version,
	)
}

func (o *Operator) handleExit(req *fixpacks.ReqOpts, err error) {
	if err != nil {
		log.Errorf("fixpacks: failed to %s %s(%v)", req.Status.Desired, req.Version, err)
		req.SetError()
	} else {
		log.Infof("fixpacks: %s %s successfully", req.Status.Desired, req.Version)
		req.SetCompleted()
	}

	o.reportToController(req)
}

func (o *Operator) reportToController(req *fixpacks.ReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("fixpacks: failed to report %s result to control(%v)", req.Version, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateFixpackTaskUrl())
	if err != nil {
		log.Errorf("fixpacks: failed to send fixpack(%s) task update to %s(%v)", req.Version, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"fixpacks: has error response from %s fixpack %s task update(%v)",
			node.Hostname,
			req.Version,
			string(resp.Body()),
		)
	}
}
