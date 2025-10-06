package integrations

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "getStorage":
		return h.parseGetStorageParams()
	case "verifyStorage":
		return h.parseVerifyStorageParams()
	case "createStorage":
		return h.parseCreateStorageParams()
	case "updateStorage":
		return h.parseUpdateStorageParams()
	case "setStorageAsDefault":
		return h.parseSetDefaultStorageParams()
	case "deleteStorage":
		return h.parseDeleteStorageParams()
	case "updateStorageTask":
		return h.parseUpdateStorageTaskOptions()
	case "createStorageModel":
		return h.parseCreateStorageModelParams()
	case "updateStorageModel":
		return h.parseUpdateStorageModelParams()
	case "updateAllStorageModels":
		return h.parseUpdateAllStorageModelParams()
	case "deleteStorageModel":
		return h.parseDeleteStorageModelParams()
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

func (h *helper) parseVerifyStorageParams() error {
	err := h.c.ShouldBindJSON(&h.storageReqOpts.CinderDetails)
	if err != nil {
		return err
	}

	if h.storageReqOpts.Name == "" {
		return errors.New("storage name is required")
	}

	h.storageReqOpts.ReqId = h.reqId
	h.storageReqOpts.Hostname = base.Hostname
	return nil
}

func (h *helper) parseCreateStorageParams() error {
	err := h.c.ShouldBindJSON(&h.storageReqOpts.CinderDetails)
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
	err := h.c.ShouldBindJSON(&h.storageReqOpts.CinderDetails)
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

func (h *helper) parseSetDefaultStorageParams() error {
	h.storageReqOpts.Name = h.c.Param("storageName")
	if h.storageReqOpts.Name == "" {
		return errors.New("storage name is required")
	}

	h.storageReqOpts.ReqId = h.reqId
	h.storageReqOpts.Hostname = base.Hostname
	h.storageReqOpts.SetSettingAsDefault()
	return nil
}

func (h *helper) parseDeleteStorageParams() error {
	h.storageReqOpts.Name = h.c.Param("storageName")
	if h.storageReqOpts.Name == "" {
		return errors.New("storage name is required")
	}

	h.storageReqOpts.ReqId = h.reqId
	h.storageReqOpts.Hostname = base.Hostname
	h.storageReqOpts.SetDeleting()
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

func (h *helper) parseCreateStorageModelParams() error {
	err := h.loadStorageModel()
	if err != nil {
		return fmt.Errorf("failed to load storage model req(%v)", err)
	}

	if h.modelReqOpts.Driver == "" {
		return errors.New("driver is required")
	}

	h.modelReqOpts.ReqId = h.reqId
	h.modelReqOpts.Hostname = base.Hostname
	h.modelReqOpts.SetCreating()
	return nil
}

func (h *helper) parseUpdateStorageModelParams() error {
	err := h.loadStorageModel()
	if err != nil {
		return fmt.Errorf("failed to load storage model req(%v)", err)
	}

	if h.modelReqOpts.Driver == "" {
		return errors.New("driver is required")
	}

	h.modelReqOpts.ReqId = h.reqId
	h.modelReqOpts.Hostname = base.Hostname
	h.modelReqOpts.SetUpdating()
	return nil
}

func (h *helper) parseUpdateAllStorageModelParams() error {
	err := h.loadStorageModelList()
	if err != nil {
		return fmt.Errorf("failed to load storage model list(%v)", err)
	}

	for _, reqOpts := range h.batchModelReqOpts {
		b, err := json.Marshal(reqOpts)
		if err != nil {
			return fmt.Errorf("failed to marshal storage model options(%v)", err)
		}

		if reqOpts.Driver == "" {
			return fmt.Errorf("has empty driver in the %s", string(b))
		}
	}

	return nil
}

func (h *helper) parseDeleteStorageModelParams() error {
	h.modelReqOpts.Driver = h.c.Param("driverName")
	if h.modelReqOpts.Driver == "" {
		return errors.New("driver is required")
	}

	h.modelReqOpts.ReqId = h.reqId
	h.modelReqOpts.Hostname = base.Hostname
	h.modelReqOpts.SetDeleting()
	return nil
}
