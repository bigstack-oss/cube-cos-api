package integrations

import "fmt"

func (h *helper) checkTaskUpdateReq() error {
	if h.storageReqOpts.Name == "" {
		return fmt.Errorf("storage name is required")
	}

	if h.storageReqOpts.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	return nil
}
