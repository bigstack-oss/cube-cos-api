package tunings

import (
	"fmt"
	"io"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func (h *helper) decodeTuningReq(reqBody io.ReadCloser) (*tunings.Tuning, error) {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		log.Errorf("tunings(%s): failed to read request body: %v", h.reqId, err)
		return nil, err
	}

	tuning := &tunings.Tuning{}
	err = json.Unmarshal(b, tuning)
	if err != nil {
		log.Errorf("tunings(%s): failed to decode tuning request: %v", h.reqId, err)
		return nil, fmt.Errorf("the request body is brought or not valid")
	}

	return tuning, nil
}

func (h *helper) checkTuningPatchReq() error {
	err := tunings.CheckSpec(h.tuning)
	if err != nil {
		return err
	}

	return h.checkHostsAreValid()
}

func (h *helper) checkTuningResetReq() error {
	return h.checkHostsAreValid()
}

func (h *helper) checkHostsAreValid() error {
	for _, host := range h.tuning.Hosts {
		_, err := nodes.Get(host.Name)
		if err != nil {
			return fmt.Errorf("host %s not found", host.Name)
		}
	}

	return nil
}

func (h *helper) checkTaskUpdateReq(tuning *tunings.Tuning) error {
	if tuning.Id == "" {
		return fmt.Errorf("tuning id is required")
	}

	if tuning.Status == nil {
		return fmt.Errorf("tuning status is required")
	}

	return nil
}
