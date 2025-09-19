package images

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operate(req *images.ReqOpts) error {
	if req.Status == nil {
		return fmt.Errorf("status is required for image request")
	}

	switch req.Status.Desired {
	case status.Imported:
		return o.importImage(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for image %s",
		req.Status.Desired,
		req.Name,
	)
}

func (o *Operator) importImage(req *images.ReqOpts) error {
	opts, err := cubecos.GenCreateOptsByReqOpts(*req)
	if err != nil {
		return err
	}

	go o.syncProgressToController(req, &opts.StreamingLogs)
	err = cubecos.ImportImage(opts)
	if err != nil {
		return err
	}

	updateOpts := o.genImageCustomProperties(req)
	cubecos.SetImagePropertiesByName(req.Name, updateOpts)
	return nil
}

func (o *Operator) genImageCustomProperties(req *images.ReqOpts) opsimage.UpdateOpts {
	return opsimage.UpdateOpts{
		opsimage.UpdateImageProperty{
			Op:    opsimage.AddOp,
			Name:  images.CubeDefinedOs,
			Value: req.Os,
		},
		opsimage.UpdateImageProperty{
			Op:    opsimage.AddOp,
			Name:  images.DefaultOsDistro,
			Value: req.Os,
		},
		opsimage.UpdateImageProperty{
			Op:    opsimage.AddOp,
			Name:  images.CubeDefinedDestination,
			Value: req.Destination,
		},
	}
}

func (o *Operator) syncProgressToController(req *images.ReqOpts, streamingLogs *chan float64) {
	if streamingLogs == nil {
		return
	}

	for {
		progress, ok := <-*streamingLogs
		if !ok {
			return
		}

		req.Status.ProcessPercent = progress
		o.reportToController(req)
	}
}

func (o *Operator) handleExit(req *images.ReqOpts, err error) {
	if err != nil {
		log.Errorf("images: failed to %s %s(%v)", req.Status.Desired, req.Name, err)
		req.SetError()
	} else {
		log.Infof("images: %s %s successfully", req.Status.Desired, req.Name)
		req.SetCompleted()
	}

	o.reportToController(req)
}

func (o *Operator) reportToController(req *images.ReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("images: failed to report %s result to control(%v)", req.Name, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateImageTaskUrl())
	if err != nil {
		log.Errorf("images: failed to send image(%s) task update to %s(%v)", req.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"images: has error response from %s image %s task update(%v)",
			node.Hostname,
			req.Name,
			string(resp.Body()),
		)
	}
}
