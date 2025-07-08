package triggers

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"

func (h *helper) setCreationReq() {
	req := triggers.ApiSchema{}
	req.Name = h.applyOpts.Name
	req.Description = h.applyOpts.Description
	req.SetMatchRule(h.applyOpts.ApplyAttributes)
	req.SetResponses(h.applyOpts.ApplyResponse)
	req.IsReportRequired = true
	h.trigger = req
}

func (h *helper) setUpdateReq() {
	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.Match = h.trigger.GenMatchRule()
	h.trigger.SetUpdating()
	h.trigger.IsReportRequired = h.isClusterWiseRequired
	h.setResponseTypes()

	trigger, found := triggers.Get(h.trigger.Name)
	if found {
		h.trigger.Enabled = trigger.Enabled
	}

	for i := range h.trigger.Response.Emails {
		h.trigger.Response.Emails[i].Enabled = true
	}

	for i := range h.trigger.Response.Slacks {
		h.trigger.Response.Slacks[i].Enabled = true
	}
}

func (h *helper) setDeletionReq() {
	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.SetDeleting()
	h.trigger.IsReportRequired = h.isClusterWiseRequired
}
