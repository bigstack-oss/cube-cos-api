package images

import (
	"fmt"
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	log "go-micro.dev/v5/logger"
)

func (h *helper) validateValues() error {
	if h.isImageExist() {
		err := fmt.Errorf("image %s already exists", h.reqOpts.Name)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if !h.isProjectExists() {
		err := fmt.Errorf("invalid project %s", h.reqOpts.Project)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if !h.isDomainExists() {
		err := fmt.Errorf("invalid domain %s", h.reqOpts.Domain)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if !h.isVisibilityValid() {
		err := fmt.Errorf("invalid visibility %s", h.reqOpts.Visibility)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) isImageExist() bool {
	isExists, err := h.openstack.IsImageExist(h.reqOpts.Name)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isProjectExists() bool {
	isExists, err := h.openstack.IsProjectExists(h.reqOpts.Project)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isDomainExists() bool {
	isExists, err := h.openstack.IsDomainExists(h.reqOpts.Domain)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isVisibilityValid() bool {
	return slices.Contains(images.Visibilitise, h.reqOpts.Visibility)
}
