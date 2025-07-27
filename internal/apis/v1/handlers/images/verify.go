package images

import (
	"fmt"
	"slices"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	log "go-micro.dev/v5/logger"
)

func (h *helper) validateReq() error {
	err := h.checkEmptyValues()
	if err != nil {
		return err
	}

	return h.validateValues()
}

func (h *helper) checkEmptyValues() error {
	if h.reqOpts.File == "" {
		err := fmt.Errorf("file is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Name == "" {
		err := fmt.Errorf("name is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Project == "" {
		err := fmt.Errorf("project is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Domain == "" {
		err := fmt.Errorf("domain is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Os == "" {
		err := fmt.Errorf("os is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Destination == "" {
		err := fmt.Errorf("destination is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Visibility == "" {
		err := fmt.Errorf("visibility is required for import image")
		log.Errorf("images(%s): %v", h.reqId, err)
	}

	return nil
}

func (h *helper) validateValues() error {
	if h.isImageValid() {
		err := fmt.Errorf("image %s already exists", h.reqOpts.Name)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.isProjectValid() {
		err := fmt.Errorf("invalid project %s", h.reqOpts.Project)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.isDomainValid() {
		err := fmt.Errorf("invalid domain %s", h.reqOpts.Domain)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	if h.isVisibilityValid() {
		err := fmt.Errorf("invalid visibility %s", h.reqOpts.Visibility)
		log.Errorf("images(%s): %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) isImageValid() bool {
	isExists, err := h.openstack.IsImageExist(h.reqOpts.Name)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isProjectValid() bool {
	isExists, err := h.openstack.IsProjectExists(h.reqOpts.Project)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isDomainValid() bool {
	isExists, err := h.openstack.IsDomainExists(h.reqOpts.Domain)
	if err != nil {
		return false
	}

	return isExists
}

func (h *helper) isVisibilityValid() bool {
	return slices.Contains(images.Visibilitise, h.reqOpts.Visibility)
}
