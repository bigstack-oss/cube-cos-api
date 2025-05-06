package healths

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/api/query"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	handler string

	serviceType string
	moduleType  string
	module      *v1.Module

	period *v1.Period
	past   string

	watch bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	var err error

	switch h.handler {
	case "getHealthSummary":
		err = h.parseSummaryParams()
	case "genServiceHealthHistory":
		err = h.parseServiceHealthParams()
	case "getModuleHealthHistory":
		err = h.parseModuleHealthParams()
	case "forceRepairModule":
		err = h.parseModuleRepairParams()
	}
	if err != nil {
		return nil, err
	}

	return h, nil
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
	h.module, err = h.parseModule()
	if err != nil {
		return err
	}

	return nil
}
