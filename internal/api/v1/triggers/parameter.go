package triggers

func (h *helper) getTriggerName() string {
	return h.c.Param("triggerName")
}
