package triggers

import (
	"errors"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listTriggers":
		return h.parseListParams()
	case "verifyMaterialScript":
		return h.parseMaterialVerifyParams()
	case "createTrigger":
		return h.parseCreateParams()
	case "updateTrigger":
		return h.parseUpdateParams()
	case "deleteTrigger":
		return h.parseDeleteparams()
	case "enableOrDisableTrigger":
		return h.parseToggleParams()
	case "updateTriggerTask":
		return h.parseTaskParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseMaterialVerifyParams() error {
	err := h.c.ShouldBindJSON(&h.trigger)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseCreateParams() error {
	err := h.c.ShouldBindJSON(&h.applyOpts)
	if err != nil {
		return err
	}

	if h.applyOpts.Name == "" {
		return errors.New("trigger name is required")
	}

	return h.checkMaterials()
}

func (h *helper) parseUpdateParams() error {
	err := h.parseTrigger()
	if err != nil {
		return err
	}

	h.isClusterWiseRequired = queries.ParseClusterWise(h.c)
	return nil
}

func (h *helper) parseDeleteparams() error {
	name := h.c.Param("triggerName")
	if !cubecos.IsTriggerExist(name) {
		return errors.New("trigger does not exist")
	}

	return nil
}

func (h *helper) parseToggleParams() error {
	name := h.c.Param("triggerName")
	if !cubecos.IsTriggerExist(name) {
		return errors.New("trigger does not exist")
	}

	h.isClusterWiseRequired = queries.ParseClusterWise(h.c)
	return nil
}

func (h *helper) parseTaskParams() error {
	return h.parseTrigger()
}

func (h *helper) parseTriggerEnablement() error {
	err := h.c.ShouldBindJSON(&h.toggle)
	if err != nil {
		return err
	}

	name := h.c.Param("triggerName")
	trigger, found := triggers.Get(name)
	if !found {
		return fmt.Errorf("trigger %s not found", h.trigger.Name)
	}

	h.trigger = *trigger
	h.trigger.IsReportRequired = h.isClusterWiseRequired
	h.trigger.Enabled = h.toggle.Enable
	h.trigger.SetUpdating()
	return nil
}
