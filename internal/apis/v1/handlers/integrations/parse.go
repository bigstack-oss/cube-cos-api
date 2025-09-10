package integrations

import "errors"

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "getStorage":
		return h.parseGetStorageParams()
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
