package triggers

func (h *helper) syncInProgressInfo(trigger *triggerResp) {
	trigger.SetOk()
	if !h.hasUpdatingRecord(*trigger) {
		return
	}

	updating, err := h.getUpdatingRecord(*trigger)
	if err != nil {
		return
	}

	h.syncUpdatingTrigger(trigger, updating)
}

func (h *helper) syncUpdatingTrigger(trigger, updating *triggerResp) {
	trigger.Attribute = updating.Attribute
	trigger.Response = updating.Response
	trigger.Enabled = updating.Enabled
	trigger.Description = updating.Description
	trigger.Status.IsProcessing = updating.Status.IsProcessing
	trigger.Status.Current = updating.Status.Current
	trigger.Status.UpdatedAt = updating.Status.UpdatedAt
}
