package triggers

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"

func (h *helper) setCreationReq() {
	req := triggers.ApiSchema{Attributes: h.applyOpts.Attributes}
	req.Name = h.applyOpts.Name
	req.Description = h.applyOpts.Description
	req.Topic = "events"
	req.SetMatchRule()
	req.SetResponses(h.applyOpts.ApplyResponse)
	req.Enabled = true
	req.IsReportRequired = true
	h.trigger = req
}

func (h *helper) setUpdateReq() {
	req := triggers.ApiSchema{Attributes: h.applyOpts.Attributes}
	req.Name = h.c.Param("triggerName")
	req.Description = h.applyOpts.Description
	req.SetMatchRule()
	req.SetResponses(h.applyOpts.ApplyResponse)
	req.SetUpdating()
	h.trigger.IsReportRequired = h.isClusterWiseRequired
	trigger, found := triggers.Get(h.trigger.Name)
	if found {
		h.trigger.Enabled = trigger.Enabled
	}

	h.trigger = req
}

func (h *helper) setDeletionReq() {
	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.SetDeleting()
	h.trigger.IsReportRequired = h.isClusterWiseRequired
}
