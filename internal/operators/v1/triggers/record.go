package triggers

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(trigger trigger.Options, err error) {
	if err != nil {
		log.Errorf("trigger: failed to %s %s: %s", trigger.Status.Desired, trigger.Name, err.Error())
		trigger.SetError()
	} else {
		log.Infof("trigger: %s %s successfully", trigger.Status.Desired, trigger.Name)
		trigger.SetCompleted()
	}

	err = o.reportToController(trigger)
	if err != nil {
		return
	}
}

func (o *Operator) reportToController(trigger trigger.Options) error {
	node, err := definition.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("trigger: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeader(node.GenAuthHeader()).
		SetBody(trigger.GenTaskUpdate()).
		Patch(node.PatchTriggerTaskUrl(trigger))
	if err != nil {
		log.Errorf("failed to send trigger %s to %s: %s", trigger.Name, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to send trigger %s to %s: %d %s", trigger.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
