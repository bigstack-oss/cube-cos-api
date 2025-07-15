package node

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateDevice(req nodes.DeviceReqOpts) error {
	switch req.Status.Desired {
	case status.Added:
		return cubecos.AddDevice(req)
	case status.Removed:
		return cubecos.RemoveDevice(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for node device(%s)",
		req.Status.Desired,
		req.Device,
	)
}

func (o *Operator) handleDeviceExit(req nodes.DeviceReqOpts, err error) {
	if err != nil {
		log.Errorf("nodes: failed to %s %s(%v)", req.Status.Desired, req.Device, err)
		req.SetError()
	} else {
		log.Infof("nodes: %s %s successfully", req.Status.Desired, req.Device)
		req.SetCompleted()
	}

	o.reportToController(req)
}

func (o *Operator) reportToController(req nodes.DeviceReqOpts) {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("nodes: failed to get controller nodes(%v)", err)
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.PatchDeviceTaskUrl())
	if err != nil {
		log.Errorf("nodes: failed to send device(%s) task update to %s(%v)", req.Device, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"nodes: has error response from %s device task update(%d %v)",
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
	}
}
