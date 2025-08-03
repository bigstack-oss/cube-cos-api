package volumes

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateImageConvertion(req *images.ReqOpts) error {
	if req.Status == nil {
		return fmt.Errorf("status is required for image convertion request")
	}

	switch req.Status.Desired {
	case status.Imported:
		return o.convertImage(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for converting image %s",
		req.Status.Desired,
		req.Name,
	)
}

func (o *Operator) convertImage(req *images.ReqOpts) error {
	opts := req.GenCreateOpts()
	go o.syncProgressToController(req, &opts.StreamingLogs)
	return cubecos.ImportImage(&opts)
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
		o.reportVolumeConvertionTaskToController(req)
	}
}

func (o *Operator) handleExit(req *images.ReqOpts, err error) {
	if err != nil {
		log.Errorf("volumes: failed to %s %s(%v)", req.Status.Desired, req.Name, err)
		req.SetError()
	} else {
		log.Infof("volumes: %s %s successfully", req.Status.Desired, req.Name)
		req.SetCompleted()
	}

	o.reportVolumeConvertionTaskToController(req)
}

func (o *Operator) reportVolumeConvertionTaskToController(req *images.ReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("volumes: failed to report %s result to control(%v)", req.Name, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.UpdateVolumeImageTaskUrl())
	if err != nil {
		log.Errorf("volumes: failed to send image convertion(%s) task update to %s(%v)", req.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"volumes: has error response from %s image convertion %s task update(%v)",
			node.Hostname,
			req.Name,
			string(resp.Body()),
		)
	}
}
