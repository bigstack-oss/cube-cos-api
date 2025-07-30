package triggers

import (
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
