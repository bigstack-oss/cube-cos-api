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
		return cubecos.UpdateStorageModel(req)
	case status.Deleted:
		return cubecos.DeleteStorageModel(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for model(%s %s)",
		req.Status.Desired,
		req.Vendor,
		req.Product,
	)
}

func (o *Operator) handleModelExit(req storages.ModelReqOpts, err error) {
	if err != nil {
		log.Errorf("storages: failed to %s %s %s(%v)", req.Status.Desired, req.Vendor, req.Model, err)
		req.SetFailed()
	} else {
		log.Infof("storages: %s %s %s successfully", req.Status.Desired, req.Vendor, req.Model)
		req.SetCompleted()
	}

	o.reportModelToController(req)
}

func (o *Operator) reportModelToController(req storages.ModelReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("storages: failed to report %s %s result to control(%v)", req.Vendor, req.Model, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateModelTaskUrl())
	if err != nil {
		log.Errorf("storages: failed to send model %s %s to %s(%v)", req.Vendor, req.Model, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"storages: failed to send model %s %s to %s(%s)",
			req.Vendor,
			req.Model,
			node.Hostname,
			string(resp.Body()),
		)
	}
}
