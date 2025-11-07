package integrations

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

func (h *helper) isVerificationReq() bool {
	return h.storageReqOpts.Status.Current == status.Verified
}

func (h *helper) checkStorageTaskUpdateReq() error {
	if h.storageReqOpts.CinderDetails.Name == "" {
		return fmt.Errorf("storage name is required")
	}

	if h.storageReqOpts.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	return nil
}

func (h *helper) checkModelTaskUpdateReq() error {
	if h.modelReqOpts.Model.Driver == "" {
		return fmt.Errorf("model driver is required")
	}

	if h.modelReqOpts.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	return nil
}
