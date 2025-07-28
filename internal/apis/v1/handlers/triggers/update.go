package triggers

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"

func (h *helper) syncUpdatingInfo(trigger *triggers.ApiSchema) {
	trigger.SetOk()
	if !h.hasUpdateHistory(*trigger) {
		return
	}

	record, err := h.getUpdateRecord(*trigger)
	if err != nil {
		return
	}

	h.syncUpdatingPayload(trigger, record)
	h.syncUpdatingStatus(trigger, record)
}

func (h *helper) syncUpdatingPayload(trigger *triggers.ApiSchema, record *triggers.ApiSchema) {
	trigger.Attribute = record.Attribute
	trigger.Types = record.Types
	trigger.Response = record.Response
	trigger.Enabled = record.Enabled
	trigger.Description = record.Description
}

func (h *helper) syncUpdatingStatus(trigger *triggers.ApiSchema, record *triggers.ApiSchema) {
	trigger.Status.IsUpdating = record.Status.IsUpdating
	trigger.Status.Current = record.Status.Current
	trigger.Status.UpdatedAt = record.Status.UpdatedAt
}
