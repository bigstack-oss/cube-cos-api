package integrations

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) checkTaskUpdateReq() error {
	if h.storageReqOpts.Name == "" {
		return fmt.Errorf("storage name is required")
	}

	if h.storageReqOpts.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	return nil
}

func (h *helper) checkIfStorageIsDefaulted() error {
	storage, err := cubecos.GetStorage(h.storageReqOpts.Name)
	if err != nil {
		bodies.SetInternalServerError(h.c, err)
		return err
	}

	if storage.AsDefault {
		err := fmt.Errorf("the %s is already the default storage", h.storageReqOpts.Name)
		bodies.SetBadRequest(h.c, err, nil)
		return err
	}

	return nil
}
