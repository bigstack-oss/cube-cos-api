package triggers

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(req triggers.ReqOpts, err error) {
	if err != nil {
		log.Errorf("triggers: failed to %s %s(%v)", req.Status.Desired, req.Name, err)
		req.SetError()
	} else {
		log.Infof("triggers: %s %s successfully", req.Status.Desired, req.Name)
		req.SetCompleted()
	}

	o.reportToController(req)
}

func (o *Operator) reportToController(req triggers.ReqOpts) {
	node, err := cubecos.GetVirtualIpController()
	if err != nil {
		log.Errorf("triggers: failed to report %s result to control(%v)", req.Name, err)
		return
	}

	resp, err := o.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(req).
		Patch(node.PatchTriggerTaskUrl())
	if err != nil {
		log.Errorf("triggers: failed to send trigger %s to %s(%v)", req.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"triggers: failed to send trigger %s to %s(%s)",
			req.Name,
			node.Hostname,
			string(resp.Body()),
		)
	}
}
