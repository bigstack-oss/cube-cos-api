package triggers

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/google/uuid"
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
	h.trigger.Id = uuid.New().String()
	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.GenMatchRule()
	h.trigger.InitStatus("updating", "update")
}

func (h *helper) parseTaskId() error {
	h.trigger.Id = h.c.Param("taskId")
	if h.trigger.Id == "" {
		return errors.New("task id is required")
	}

	return nil
}
