package node

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateOsd(req nodes.OsdReqOpts) error {
	switch req.Status.Desired {
	case status.Restarted:
		return o.restartOsd(req)
	case status.Reweighted:
		return cubecos.ReweightOsd(req)
	case status.Removed:
		return cubecos.RemoveOsd(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for node osd(%s)",
		req.Status.Desired,
		req.OsdId,
	)
}

func (o *Operator) restartOsd(req nodes.OsdReqOpts) error {
	err := cubecos.RestartOsd(req)
	if err != nil {
		return err
	}

	return cubecos.WaitOsdStatus(
		req.OsdId,
		status.Up,
		300,
	)
}

func (o *Operator) handleOsdExit(req nodes.OsdReqOpts, err error) {
	if err != nil {
		log.Errorf("nodes: failed to %s %s(%v)", req.Status.Desired, req.OsdId, err)
		req.SetError()
	} else {
		log.Infof("nodes: %s %s successfully", req.Status.Desired, req.OsdId)
		req.SetCompleted()
	}

	o.reportOsdToController(req)
}

func (o *Operator) reportOsdToController(req nodes.OsdReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("nodes: %v", err)
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateOsdTaskUrl())
	if err != nil {
		log.Errorf("nodes: failed to send osd(%s) task update to %s(%v)", req.OsdId, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"nodes: has error response from %s %s task update(%v)",
			node.Hostname,
			req.OsdId,
			string(resp.Body()),
		)
	}
}
