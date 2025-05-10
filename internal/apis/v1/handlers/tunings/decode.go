package tunings

import (
	"fmt"
	"io"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	json "github.com/json-iterator/go"
	log "go-micro.dev/v5/logger"
)

func (h *helper) decodeTuningReq(reqBody io.ReadCloser) (*v1.Tuning, error) {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		log.Errorf("request(%s): failed to read request body: %s", queries.GetReqId(h.c), err.Error())
		return nil, err
	}

	tuning := &v1.Tuning{}
	err = json.Unmarshal(b, tuning)
	if err != nil {
		log.Errorf("request(%s): failed to decode tuning request: %s", queries.GetReqId(h.c), err.Error())
		return nil, fmt.Errorf("the request body is brought or not valid")
	}

	return tuning, nil
}

func (h *helper) checkTuningPatchReq() error {
	err := v1.CheckTuningSpec(h.tuning)
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
			return fmt.Errorf("host(%s) not found", host.Name)
		}
	}

	return nil
}

func (h *helper) checkTaskUpdateReq(tuning *v1.Tuning) error {
	if tuning.Id == "" {
		return fmt.Errorf("tuning id is required")
	}

	if tuning.Status == nil {
		return fmt.Errorf("tuning status is required")
	}

	return nil
}
