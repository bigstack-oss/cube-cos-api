package triggers

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
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
