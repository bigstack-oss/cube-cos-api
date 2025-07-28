package triggers

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (h *helper) setCreationReq() {
	// req := triggers.ApiSchema{Attribute: h.reqOpts.Attribute}
	// req.Name = h.reqOpts.Name
	// req.Description = h.reqOpts.Description
	// req.Topic = "events"
	// req.SetMatchRule()
	// req.SetResponses(h.reqOpts.ReqResponse)
	// req.Enabled = true
	// req.SetUpdating()
	// h.trigger.IsReportRequired = h.requireClusterUpdate
	// h.trigger = req
}

func (h *helper) genResponseByReq() triggers.Responses {
	resp := triggers.Responses{
		Emails: h.reqOpts.ReqResponse.Emails,
		Slacks: h.reqOpts.ReqResponse.Slacks,
		Execs:  triggers.Execs{Shells: []string{}},
	}

	if h.reqOpts.ReqResponse.Script.Name != "" {
		resp.Execs.Shells = append(
			resp.Execs.Shells,
			fmt.Sprintf("%s.shell", h.reqOpts.ReqResponse.Script.Name),
		)
	}

	return resp
}

func (h *helper) setUpdateReq() {
	req := triggers.ApiSchema{Attribute: h.reqOpts.Attribute}
	req.Name = h.c.Param("triggerName")
	req.Description = h.reqOpts.Description
	req.Topic = "events"
	req.SetMatchRule()
	req.SetResponses(h.reqOpts.ReqResponse)
	req.SetUpdating()
	// h.reqOpts.IsReportRequired = h.requireClusterUpdate
	trigger, found := triggers.Get(h.reqOpts.Name)
	if found {
		h.reqOpts.Enabled = trigger.Enabled
	}

	// h.trigger = req
}

func (h *helper) setDeletionReq() {
	h.reqOpts.Name = h.c.Param("triggerName")
	h.reqOpts.SetDeleting()
	// h.trigger.IsReportRequired = h.requireClusterUpdate
}
