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

	o.reportToController(tuning)
}

func (o *Operator) reportToController(tuning tunings.Tuning) error {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("tunings: failed to get controller nodes: %v", err)
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(tuning.GenTaskUpdate()).
		Patch(node.PatchTuningTaskUrl())
	if err != nil {
		log.Errorf("tunings: failed to update %s to %s: %v", tuning.Name, node.Hostname, err)
		return err
	}

	if resp.IsError() {
		log.Errorf("tunings: has resp error from %s: %s(%s)", node.Hostname, tuning.Name, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
