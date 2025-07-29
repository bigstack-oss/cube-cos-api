package triggers

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (h *helper) parseTriggerName() string {
	name := h.c.Param("triggerName")
	builtInMap := triggers.GetBuiltInNameMap()

	builtInName, found := builtInMap[name]
	if found {
		return builtInName
	}

	return name
}

func (h *helper) parseTrigger() error {
	name := h.parseTriggerName()
	if !cubecos.IsTriggerExist(name) {
		return errors.New("trigger does not exist")
	}

	err := h.c.ShouldBindJSON(&h.reqOpts)
	if err != nil {
		return err
	}

	h.reqOpts.Name = name
	return nil
}
