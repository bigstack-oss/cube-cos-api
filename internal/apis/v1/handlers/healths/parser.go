package healths

import (
	"fmt"

	query "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/services"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "getHealthSummary":
		return h.parseSummaryParams()
	case "genServiceHealthHistory":
		return h.parseServiceHealthParams()
	case "getModuleHealthHistory":
		return h.parseModuleHealthParams()
	case "forceRepairModule":
		return h.parseModuleRepairParams()
	case "deleteModuleRepairTask":
		return h.parseForceRepairTaskParams()
	}

	return nil
}

func (h *helper) parseSummaryParams() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseServiceHealthParams() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}

	h.period, err = query.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.serviceType = h.c.Param("serviceType")
	if !cubecos.IsValidService(h.serviceType) {
		return fmt.Errorf("invalid serviceType: %s", h.serviceType)
	}

	return nil
}

func (h *helper) parseModuleHealthParams() error {
	var err error
	h.watch, err = query.GetWatch(h.c)
	if err != nil {
		return err
	}

	h.past, err = query.GetPast(h.c)
	if err != nil {
		return err
	}

	h.period, err = query.GetPeriod(h.c)
	if err != nil {
		return err
	}

	h.serviceType = h.c.Param("serviceType")
	if !cubecos.IsValidService(h.serviceType) {
		return fmt.Errorf("invalid serviceType: %s", h.serviceType)
	}

	h.moduleType = h.c.Param("moduleType")
	if !cubecos.IsValidServiceAndModule(h.serviceType, h.moduleType) {
		return fmt.Errorf("invalid serviceType' %s' or module '%s'", h.serviceType, h.moduleType)
	}

	return nil
}

func (h *helper) parseModuleRepairParams() error {
	var err error
	h.moduleType = h.c.Param("moduleType")
	h.module, err = h.parseModule()
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseModule() (*services.Module, error) {
	m := h.c.Param("moduleType")
	module, found := cubecos.Modules[m]
	if !found {
		return nil, fmt.Errorf("module(%s) not found", m)
	}

	return &module, nil
}

func (h *helper) parseForceRepairTaskParams() error {
	h.moduleType = h.c.Param("moduleType")
	if h.moduleType == "" {
		return fmt.Errorf("moduleType is required")
	}

	return nil
}
