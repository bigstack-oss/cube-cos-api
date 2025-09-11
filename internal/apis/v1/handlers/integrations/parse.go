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
	case "createStorage":
		return h.parseCreateStorageParams()
	case "updateStorage":
		return h.parseUpdateStorageParams()
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
	err := h.c.ShouldBindYAML(&h.modelReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse create storage model options(%v)",
			err,
		)
	}

	if h.modelReqOpts.Vendor == "" {
		return errors.New("vendor is required")
	}

	if h.modelReqOpts.Product == "" {
		return errors.New("product is required")
	}

	h.modelReqOpts.ReqId = h.reqId
	h.modelReqOpts.Hostname = base.Hostname
	h.modelReqOpts.SetCreating()
	return nil
}

func (h *helper) parseUpdateStorageModelParams() error {
	err := h.c.ShouldBindYAML(&h.modelReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse update storage model options(%v)",
			err,
		)
	}

	if h.modelReqOpts.Vendor == "" {
		return errors.New("vendor is required")
	}

	if h.modelReqOpts.Product == "" {
		return errors.New("product is required")
	}

	h.modelReqOpts.ReqId = h.reqId
	h.modelReqOpts.Hostname = base.Hostname
	h.modelReqOpts.SetUpdating()
	return nil
}

func (h *helper) parseUpdateAllStorageModelParams() error {
	err := h.c.ShouldBindYAML(&h.batchModelReqOpts)
	if err != nil {
		return fmt.Errorf(
			"failed to parse batch storage model options(%v)",
			err,
		)
	}

	for _, reqOpts := range h.batchModelReqOpts {
		b, err := json.Marshal(reqOpts)
		if err != nil {
			return fmt.Errorf("failed to marshal storage model options(%v)", err)
		}

		if reqOpts.Vendor == "" {
			return fmt.Errorf("has empty vendor in the %s", string(b))
		}

		if reqOpts.Product == "" {
			return fmt.Errorf("has empty product in the %s", string(b))
		}
	}

	return nil
}

func (h *helper) parseDeleteStorageModelParams() error {
	h.modelReqOpts.Vendor = h.c.Param("vendor")
	if h.modelReqOpts.Vendor == "" {
		return errors.New("vendor is required")
	}

	h.modelReqOpts.Product = h.c.Param("product")
	if h.modelReqOpts.Product == "" {
		return errors.New("product is required")
	}

	h.modelReqOpts.ReqId = h.reqId
	h.modelReqOpts.Hostname = base.Hostname
	h.modelReqOpts.SetDeleting()
	return nil
}
