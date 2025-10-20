package storages

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateStorageReq(req storages.ReqOpts) error {
	switch req.Status.Desired {
	case status.Created, status.Updated:
		return o.updateStorage(req)
	case status.Deleted:
		return cubecos.DeleteStorage(req.CinderDetails.Name)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for storage(%s)",
		req.Status.Desired,
		req.CinderDetails.Name,
	)
}

func (o *Operator) updateStorage(req storages.ReqOpts) error {
	return cubecos.CreateStorage(req.CinderDetails)
}

func (o *Operator) handleStorageExit(req storages.ReqOpts, err error) {
	if err != nil {
		log.Errorf("storages: failed to %s %s(%v)", req.Status.Desired, req.CinderDetails.Name, err)
		req.SetFailed(err.Error())
	} else {
		log.Infof("storages: %s %s successfully", req.Status.Desired, req.CinderDetails.Name)
		req.SetCompleted()
	}

	o.reportStorageToController(req)
}

func (o *Operator) reportStorageToController(req storages.ReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("storages: failed to report %s result to control(%v)", req.CinderDetails.Name, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateStorageTaskUrl())
	if err != nil {
		log.Errorf("storages: failed to send storage %s to %s(%v)", req.CinderDetails.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"storages: failed to send storage %s to %s(%s)",
			req.CinderDetails.Name,
			node.Hostname,
			string(resp.Body()),
		)
	}
}
