package tunings

import (
	"fmt"
	"io"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func (h *helper) decodeTuningReq(reqBody io.ReadCloser) error {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		log.Errorf("tunings(%s): failed to read request body: %v", h.reqId, err)
		return err
	}

	err = json.Unmarshal(b, &h.tuning)
	if err != nil {
		log.Errorf("tunings(%s): failed to decode request: %v", h.reqId, err)
		return fmt.Errorf("the request body is not valid")
	}

	return nil
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

func (h *helper) checkTaskUpdateReq() error {
	if h.tuning.Name == "" {
		return fmt.Errorf("tuning name is required")
	}

	if h.tuning.Node == nil {
		return fmt.Errorf("node is required")
	}

	if h.tuning.Status == nil {
		return fmt.Errorf("tuning status is required")
	}

	return nil
}
