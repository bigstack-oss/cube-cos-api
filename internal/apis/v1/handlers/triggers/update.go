package triggers

func (h *helper) syncUpdatingInfo(trigger *triggerResp) {
	trigger.SetOk()
	if !h.hasUpdateHistory(*trigger) {
		return
	}

	updating, err := h.getUpdateRecord(*trigger)
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
	trigger.Status.IsUpdating = updating.Status.IsUpdating
	trigger.Status.Current = updating.Status.Current
	trigger.Status.UpdatedAt = updating.Status.UpdatedAt
}
