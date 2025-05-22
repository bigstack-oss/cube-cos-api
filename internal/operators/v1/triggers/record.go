package triggers

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(trigger triggers.ApiSchema, err error) {
	if err != nil {
		log.Errorf("triggers: failed to %s %s(%v)", trigger.Status.Desired, trigger.Name, err)
		trigger.SetError()
	} else {
		log.Infof("triggers: %s %s successfully", trigger.Status.Desired, trigger.Name)
		trigger.SetCompleted()
	}

	if trigger.IsReportRequired {
		o.reportToController(trigger)
	}
}

func (o *Operator) reportToController(trigger triggers.ApiSchema) {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("triggers: failed to get controller nodes(%v)", err)
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(trigger.GenTaskUpdate()).
		Patch(node.PatchTriggerTaskUrl(trigger))
	if err != nil {
		log.Errorf("triggers: failed to send trigger %s to %s(%v)", trigger.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"triggers: failed to send trigger %s to %s(%s)",
			trigger.Name,
			node.Hostname,
			string(resp.Body()),
		)
	}
}
