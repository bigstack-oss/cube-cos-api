package triggers

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
)

func (h *helper) syncInProgressInfo(trigger *triggerResp) {
	trigger.SetOk()
	if !h.hasUpdatingRecord(*trigger) {
		return
	}

	req, err := h.getUpdatingRecord(*trigger)
	if err != nil {
		return
	}

	if !req.Status.IsProcessing {
		h.removeFinishedRequest(req.Id)
		return
	}

	updating := h.convertReqOptsToResp(*req)
	h.syncUpdatingTrigger(trigger, updating)
}

func (h *helper) convertReqOptsToResp(req triggers.ReqOpts) *triggerResp {
	settings, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Warnf("triggers(%s): failed to get alert settings(%v)", h.reqId, err)
		return nil
	}

	emails := h.convertToEmailRecipients(req.Emails)
	slacks := h.convertToSlackChannels(req.Slacks)
	return &triggerResp{
		Name:        req.Name,
		Attribute:   req.Attribute,
		Description: req.Description,
		Enabled:     req.Enabled,
		Status:      &req.Status,
		Response: Response{
			Types:  h.getResponseTypesFromReq(req),
			Script: req.Script,
			Emails: h.parseEmailDetails(settings, emails),
			Slacks: h.parseSlackDetails(settings, slacks),
		},
	}
}

func (h *helper) syncUpdatingTrigger(trigger, updating *triggerResp) {
	if updating == nil {
		return
	}

	trigger.Attribute = updating.Attribute
	trigger.Response = updating.Response
	trigger.Enabled = updating.Enabled
	trigger.Description = updating.Description
	trigger.Status.IsProcessing = updating.Status.IsProcessing
	trigger.Status.Current = updating.Status.Current
	trigger.Status.UpdatedAt = updating.Status.UpdatedAt
}
