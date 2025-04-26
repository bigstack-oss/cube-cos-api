package tunings

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(tuning v1.Tuning, err error) {
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

func (o *Operator) reportToController(tuning v1.Tuning) error {
	node, err := v1.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("tunings: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(v1.GenNodeAuthHeaders()).
		SetBody(tuning.GenTaskUpdate()).
		Patch(node.PatchTuningTaskUrl(tuning))
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
