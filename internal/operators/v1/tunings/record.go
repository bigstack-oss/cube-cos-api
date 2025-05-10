package tunings

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(tuning tunings.Tuning, err error) {
	if err != nil {
		log.Errorf("tunings: failed to %s %s to %v(%v)", tuning.Status.Desired, tuning.Name, tuning.Value, err)
		tuning.SetError()
	} else {
		log.Infof("tunings: %s %s to %v", tuning.Status.Desired, tuning.Name, tuning.Value)
		tuning.SetCompleted()
	}

	err = o.reportToController(tuning)
	if err != nil {
		return
	}
}

func (o *Operator) reportToController(tuning tunings.Tuning) error {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("tunings: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(tuning.GenTaskUpdate()).
		Patch(node.PatchTuningTaskUrl(tuning.Id))
	if err != nil {
		log.Errorf("tunings: failed to send tuning %s to %s: %s", tuning.Name, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("tunings: failed to send tuning %s to %s: %d %s", tuning.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
