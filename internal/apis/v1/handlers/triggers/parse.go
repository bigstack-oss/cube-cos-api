package triggers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listTriggers":
		return h.parseListParams()
	case "verifyMaterialScript":
		return h.parseScriptVerifyParams()
	case "createTrigger":
		return h.parseCreateParams()
	case "updateTrigger":
		return h.parseUpdateParams()
	case "deleteTrigger":
		return h.parseDeleteparams()
	case "enableOrDisableTrigger":
		return h.parseToggleParams()
	case "updateTriggerTask":
		return h.parseTaskParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	if err != nil {
		return err
	}

	return nil
}

func (h *helper) parseScriptVerifyParams() error {
	err := h.c.ShouldBindJSON(&h.verifyScript)
	if err != nil {
		return err
	}

	script, found := h.verifyScript["script"]
	if !found {
		return errors.New("script is required for verification")
	}

	if script == "" {
		return errors.New("script cannot be empty")
	}

	h.verifyScript["script"], err = h.decodeScript(script)
	if err != nil {
		return err
	}

	return nil
}

// note:
// do not use the h.c.ShouldBindJSON() to parse the request body,
// because it will remove the content in the request body after unmarshalling.
// we need to keep the raw body for later req delegation to peer nodes.
func (h *helper) parseCreateParams() error {
	bytes := bodies.ParseReq(h.c)
	err := json.Unmarshal(bytes, &h.reqOpts)
	if err != nil {
		log.Errorf("triggers(%s): failed to parse request body: %v", h.reqId, err)
		return err
	}

	h.reqOpts.Id = h.reqId
	h.reqOpts.Enabled = true
	if h.reqOpts.Name == "" {
		return errors.New("trigger name is required")
	}

	err = h.checkMaterials()
	if err != nil {
		log.Errorf("triggers(%s): failed to check materials: %v", h.reqId, err)
		return err
	}

	h.reqOpts.SetCreating()
	return nil
}

// note:
// do not use the h.c.ShouldBindJSON() to parse the request body,
// because it will remove the content in the request body after unmarshalling.
// we need to keep the raw body for later req delegation to peer nodes.
func (h *helper) parseUpdateParams() error {
	bytes := bodies.ParseReq(h.c)
	err := json.Unmarshal(bytes, &h.reqOpts)
	if err != nil {
		log.Errorf("triggers(%s): failed to parse request body: %v", h.reqId, err)
		return err
	}

	h.reqOpts.Id = h.reqId
	h.reqOpts.Name = h.parseTriggerName()
	if h.reqOpts.Name == "" {
		return errors.New("trigger name is required")
	}

	err = h.checkMaterials()
	if err != nil {
		log.Errorf("triggers(%s): failed to check materials: %v", h.reqId, err)
		return err
	}

	h.reqOpts.SetUpdating()
	return nil
}

func (h *helper) parseDeleteparams() error {
	name := h.c.Param("triggerName")
	builtInMap := triggers.GetBuiltInNameMap()
	_, found := builtInMap[name]
	if found {
		return fmt.Errorf(
			"trigger %s is a built-in trigger and cannot be deleted",
			name,
		)
	}

	h.reqOpts.Id = h.reqId
	h.reqOpts.Name = name
	h.reqOpts.SetDeleting()
	return nil
}

func (h *helper) parseToggleParams() error {
	h.reqOpts.Name = h.parseTriggerName()
	return nil
}

func (h *helper) parseTaskParams() error {
	return h.c.ShouldBindJSON(&h.reqOpts)
}

func (h *helper) parseTriggerEnablement() error {
	err := h.c.ShouldBindJSON(&h.toggle)
	if err != nil {
		return err
	}

	name := h.parseTriggerName()
	trigger, found := triggers.Get(name)
	if !found {
		return fmt.Errorf("trigger %s not found", name)
	}

	resp := h.convertTrigger(*trigger)
	h.reqOpts.Name = resp.Name
	h.reqOpts.Enabled = h.toggle.Enable
	h.reqOpts.Description = resp.Description
	h.reqOpts.Attribute = resp.Attribute
	h.reqOpts.Response = triggers.Response{
		Script: triggers.Script{
			Name:    resp.Response.Script.Name,
			Content: resp.Response.Script.Content,
		},
		Emails: h.convertEmailStringSlice(trigger),
		Slacks: h.convertSlackStringSlice(trigger),
	}

	h.reqOpts.SetToggling()
	return nil
}

func (h *helper) convertEmailStringSlice(trigger *triggers.Trigger) []string {
	emails := []string{}
	for _, email := range trigger.Emails {
		emails = append(emails, email.Address)
	}

	return emails
}

func (h *helper) convertSlackStringSlice(trigger *triggers.Trigger) []string {
	slacks := []string{}
	for _, slack := range trigger.Slacks {
		slacks = append(slacks, slack.URL)
	}

	return slacks
}
