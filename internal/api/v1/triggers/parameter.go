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

	return h.c.ShouldBindJSON(&h.trigger)
}

func (h *helper) setUpdateInfo() {
	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.Match = h.trigger.GenMatchRule()
	h.trigger.InitUpdateStatus()
}
