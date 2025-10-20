package integrations

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

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

func (h *helper) checkIfStorageIsDefaulted() error {
	storage, err := cubecos.GetStorage(h.storageReqOpts.CinderDetails.Name)
	if err != nil {
		bodies.SetInternalServerError(h.c, err)
		return err
	}

	if storage.IsDefault {
		err := fmt.Errorf("the %s is already the default storage", h.storageReqOpts.CinderDetails.Name)
		bodies.SetBadRequest(h.c, err, nil)
		return err
	}

	return nil
}
