package tunings

import (
	"fmt"
	"io"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	json "github.com/json-iterator/go"
)

func (h *helper) decodeTuningReq(reqBody io.ReadCloser) (*definition.Tuning, error) {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		return nil, err
	}

	tuning := &definition.Tuning{}
	err = json.Unmarshal(b, tuning)
	if err != nil {
		return nil, err
	}

	return tuning, nil
}

func decodeTuningsReq(reqBody io.ReadCloser) ([]definition.Tuning, error) {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		return nil, err
	}

	tunings := []definition.Tuning{}
	err = json.Unmarshal(b, &tunings)
	if err != nil {
		return nil, err
	}

	for i := range tunings {
		tunings[i].SetNodeInfo(
			definition.CurrentRole,
			definition.AdvertiseAddr,
		)
	}

	return tunings, nil
}

func (h *helper) checkTuningPatchReq() error {
	err := definition.CheckTuningSpec(h.tuning)
	if err != nil {
		return err
	}

	return h.checkHostsAreValid()
}

func (h *helper) checkTuningResetReq() error {
	return h.checkHostsAreValid()
}

func (h *helper) checkHostsAreValid() error {
	for _, r := range h.tuning.Roles {
		for _, h := range r.Hosts {
			_, err := definition.GetNodeByHostname(h.Name)
			if err != nil {
				return fmt.Errorf("host(%s) not found", h.Name)
			}
		}
	}

	return nil
}

func (h *helper) checkTaskUpdateReq(tuning *definition.Tuning) error {
	if tuning.Id == "" {
		return fmt.Errorf("tuning id is required")
	}

	if tuning.Status == nil {
		return fmt.Errorf("tuning status is required")
	}

	return nil
}
