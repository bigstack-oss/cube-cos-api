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
