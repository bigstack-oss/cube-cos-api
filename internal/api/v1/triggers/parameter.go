package triggers

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
)

func (h *helper) getTriggerName() string {
	return h.c.Param("triggerName")
}

func (h *helper) parseTrigger() error {
	name := h.c.Param("triggerName")
	if !cubecos.IsTriggerExist(name) {
		return errors.New("trigger does not exist")
	}

	err := h.c.ShouldBindJSON(&h.trigger)
	if err != nil {
		return err
	}

	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.GenMatchRule()
	return nil
}
