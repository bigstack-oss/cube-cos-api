package volumes

import (
	"fmt"

	log "go-micro.dev/v5/logger"
)

func (h *helper) validateImageConvertionValues() error {
	if !h.isProjectExists() {
		err := fmt.Errorf("invalid project %s", h.imageReqOpts.Project)
		log.Errorf("volumes(%s): %v", h.reqId, err)
		return err
	}

	if !h.isDomainExists() {
		err := fmt.Errorf("invalid domain %s", h.imageReqOpts.Domain)
		log.Errorf("volumes(%s): %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) isProjectExists() bool {
	isExists, err := h.openstack.IsProjectExists(h.imageReqOpts.Project)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isDomainExists() bool {
	isExists, err := h.openstack.IsDomainExists(h.imageReqOpts.Domain)
	if err != nil {
		return false
	}

	return isExists
}
