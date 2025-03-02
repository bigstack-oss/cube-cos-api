package tunings

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(tuning definition.Tuning, err error) {
	if err != nil {
		log.Errorf("tuning: failed to %s %s to %v(%v)", tuning.Status.Desired, tuning.Name, tuning.Value, err)
		tuning.SetError()
	} else {
		log.Infof("tuning: %s %s to %v", tuning.Status.Desired, tuning.Name, tuning.Value)
		tuning.SetCompleted()
	}

	err = o.reportToController(tuning)
	if err != nil {
		return
	}
}

func (o *Operator) reportToController(tuning definition.Tuning) error {
	node, err := definition.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("tuning: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeader(node.GenAuthHeader()).
		SetBody(tuning.GenTaskUpdate()).
		Patch(node.PatchTuningTaskUrl(tuning))
	if err != nil {
		log.Errorf("failed to send tuning %s to %s: %s", tuning.Name, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to send tuning %s to %s: %d %s", tuning.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
