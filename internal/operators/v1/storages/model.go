package storages

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateModelReq(req storages.ModelReqOpts) error {
	switch req.Status.Desired {
	case status.Created, status.Updated:
		return o.updateModel(req)
	case status.Deleted:
		return o.deleteModel(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for model(%s)",
		req.Status.Desired,
		req.Name,
	)
}

func (o *Operator) updateModel(req storages.ModelReqOpts) error {
	return nil
}

func (o *Operator) deleteModel(req storages.ModelReqOpts) error {
	return nil
}

func (o *Operator) handleModelExit(req storages.ModelReqOpts, err error) {
	if err != nil {
		log.Errorf("storages: failed to %s %s(%v)", req.Status.Desired, req.Name, err)
		req.SetFailed()
	} else {
		log.Infof("storages: %s %s successfully", req.Status.Desired, req.Name)
		req.SetCompleted()
	}

	o.reportModelToController(req)
}

func (o *Operator) reportModelToController(req storages.ModelReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("storages: failed to report %s result to control(%v)", req.Name, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateModelTaskUrl())
	if err != nil {
		log.Errorf("storages: failed to send model %s to %s(%v)", req.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"storages: failed to send model %s to %s(%s)",
			req.Name,
			node.Hostname,
			string(resp.Body()),
		)
	}
}
