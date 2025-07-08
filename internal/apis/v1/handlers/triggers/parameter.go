package triggers

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (h *helper) parseTriggerName() string {
	return h.c.Param("triggerName")
}

func (h *helper) parseTrigger() error {
	name := h.c.Param("triggerName")
	if !cubecos.IsTriggerExist(name) {
		return errors.New("trigger does not exist")
	}

	err := h.c.ShouldBindJSON(&h.applyOpts)
	if err != nil {
		return err
	}

	h.trigger.Name = name
	return nil
}

func (h *helper) setCreationInfo() {
}

func (h *helper) setUpdateInfo() {
	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.Match = h.trigger.GenMatchRule()
	h.trigger.SetUpdating()
	h.trigger.IsReportRequired = h.isClusterWiseRequired
	h.setResponseTypes()

	trigger, found := triggers.Get(h.trigger.Name)
	if found {
		h.trigger.Enabled = trigger.Enabled
	}

	for i := range h.trigger.Response.Emails {
		h.trigger.Response.Emails[i].Enabled = true
	}

	for i := range h.trigger.Response.Slacks {
		h.trigger.Response.Slacks[i].Enabled = true
	}
}

func (h *helper) setResponseTypes() {
	if h.trigger.HasEmailRecipients() {
		h.trigger.Response.Types = append(
			h.trigger.Response.Types,
			"email",
		)
	}

	if h.trigger.HasSlackChannels() {
		h.trigger.Response.Types = append(
			h.trigger.Response.Types,
			"slack",
		)
	}
}
