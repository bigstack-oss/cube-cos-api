package triggers

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(trigger trigger.ApiOptions, err error) {
	if err != nil {
		log.Errorf("triggers: failed to %s %s: %s", trigger.Status.Desired, trigger.Name, err.Error())
		trigger.SetError()
	} else {
		log.Infof("triggers: %s %s successfully", trigger.Status.Desired, trigger.Name)
		trigger.SetCompleted()
	}

	if trigger.ShouldReportToController {
		o.reportToController(trigger)
	}
}

func (o *Operator) reportToController(trigger trigger.ApiOptions) error {
	node, err := v1.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("triggers: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(v1.GenNodeAuth()).
		SetBody(trigger.GenTaskUpdate()).
		Patch(node.PatchTriggerTaskUrl(trigger))
	if err != nil {
		log.Errorf("triggers: failed to send trigger %s to %s: %s", trigger.Name, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("triggers: failed to send trigger %s to %s: %v", trigger.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
