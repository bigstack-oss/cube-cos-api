package triggers

func (h *helper) parseTrigger() error {
	name := h.c.Param("triggerName")
	// if !v1.DoseTriggerExist(name) {
	// 	return errors.New("trigger does not exist")
	// }

	h.c.ShouldBind(&h.trigger)
	h.trigger.Name = name
	return nil
}
