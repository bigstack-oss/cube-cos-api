package integrations

import (
	"errors"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "getStorage":
		return h.parseGetStorageParams()
	case "createStorage":
		return h.parseCreateStorageParams()
	case "updateStorage":
		return h.parseUpdateStorageParams()
	case "updateStorageTask":
		return h.parseUpdateStorageTaskOptions()
	default:
		return nil
	}
}

func (h *helper) parseGetStorageParams() error {
	h.storageReqOpts.Name = h.c.Param("storageName")
	if h.storageReqOpts.Name == "" {
		return errors.New("storage name is required")
	}

	return nil
}

func (h *helper) parseCreateStorageParams() error {
	err := h.c.ShouldBindJSON(&h.storageReqOpts.Cinder)
	if err != nil {
		return err
	}

	if h.storageReqOpts.Name == "" {
		return errors.New("storage name is required")
	}

	h.storageReqOpts.ReqId = h.reqId
	h.storageReqOpts.Hostname = base.Hostname
	h.storageReqOpts.SetCreating()
	return nil
}

func (h *helper) parseUpdateStorageParams() error {
	err := h.c.ShouldBindJSON(&h.storageReqOpts.Cinder)
	if err != nil {
		return err
	}

	if h.storageReqOpts.Name == "" {
		return errors.New("storage name is required")
	}

	h.storageReqOpts.ReqId = h.reqId
	h.storageReqOpts.Hostname = base.Hostname
	h.storageReqOpts.SetUpdating()
	return nil
}

func (h *helper) parseUpdateStorageTaskOptions() error {
	err := h.c.ShouldBindJSON(&h.storageReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse patch storage task options(%v)",
			err,
		)
	}

	return nil
}
