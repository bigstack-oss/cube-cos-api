package triggers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	handler string
	trigger trigger.Options
	toggle  trigger.Toggle
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}

	switch handler {
	case "listTriggers", "getTrigger":
		return h, nil
	case "updateTrigger":
		return h.initUpdateHelper()
	case "enableOrDisableTrigger":
		return h.initToggleHelper()
	case "updateTriggerTask":
		return h.initTaskHelper()
	}

	return nil, errors.New("no internal function supported")
}

func (h *helper) initUpdateHelper() (*helper, error) {
	err := h.parseTrigger()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) initToggleHelper() (*helper, error) {
	name := h.c.Param("triggerName")
	if !cubecos.IsTriggerExist(name) {
		return nil, errors.New("trigger does not exist")
	}

	return h, nil
}

func (h *helper) initTaskHelper() (*helper, error) {
	err := h.parseTrigger()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) listTriggers() ([]trigger.Options, error) {
	triggers := []trigger.Options{}
	for _, trigger := range trigger.DefaultOptions {
		h.setResponseItemsToTrigger(&trigger)
		h.syncCubePolicy(&trigger)
		h.syncUpdateStatus(&trigger)
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

func (h *helper) syncCubePolicy(trigger *trigger.Options) {
	policy, err := cubecos.GetTriggerPolicy()
	if err != nil {
		return
	}

	policyTrigger := policy.GetTrigger(trigger.Name)
	if policyTrigger.Name == "" {
		return
	}

	h.setAttributionEnablement(trigger, policyTrigger)
	h.setResponseEnablement(trigger, policyTrigger)
	trigger.Description = policyTrigger.Description
	trigger.Enabled = policyTrigger.Enabled
}

func (h *helper) setAttributionEnablement(options *trigger.Options, policyTrigger trigger.Options) []trigger.Attribute {
	attributes := []trigger.Attribute{}
	matchRule := strings.ReplaceAll(policyTrigger.Match, `"`, ``)
	parts := strings.Split(matchRule, " OR ")

	enabledAttrs := []trigger.Attribute{}
	for _, part := range parts {
		attrPair := strings.Split(part, " == ")
		if isValidAttrPair(attrPair) {
			enabledAttrs = append(
				enabledAttrs,
				trigger.Attribute{
					Name:  strings.TrimSpace(attrPair[0]),
					Value: strings.TrimSpace(attrPair[1]),
				},
			)
		}
	}

	for i, attr := range options.Attributes {
		for _, enabledAttr := range enabledAttrs {
			if attr.Name != enabledAttr.Name {
				continue
			}

			if attr.Value != enabledAttr.Value {
				continue
			}

			options.Attributes[i].Enabled = true
			break
		}
	}

	return attributes
}

func isValidAttrPair(attrPair []string) bool {
	return len(attrPair) == 2
}

func (h *helper) setResponseEnablement(trigger *trigger.Options, policyTrigger trigger.Options) {
	for i, email := range trigger.Response.Emails {
		for _, policyEmail := range policyTrigger.Response.Emails {
			if email.Address == policyEmail.Address {
				trigger.Response.Emails[i].Enabled = true
				break
			}
		}
	}

	for i, slack := range trigger.Response.Slacks {
		for _, policySlack := range policyTrigger.Response.Slacks {
			if slack.URL == policySlack.URL {
				trigger.Response.Slacks[i].Enabled = true
				break
			}
		}
	}
}

func (h *helper) syncUpdateStatus(trigger *trigger.Options) {
	trigger.InitOkStatus()
	if !h.hasUpdateHistory(*trigger) {
		return
	}

	record, err := h.getUpdateRecord(*trigger)
	if err != nil {
		return
	}

	trigger.Status.IsUpdating = record.Status.IsUpdating
	trigger.Status.Current = record.Status.Current
	trigger.Status.UpdatedAt = record.Status.UpdatedAt
}

func (h *helper) getTrigger(name string) (*trigger.Options, error) {
	for _, trigger := range trigger.DefaultOptions {
		if trigger.Name == name {
			h.setResponseItemsToTrigger(&trigger)
			return &trigger, nil
		}
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func (h *helper) delegateTriggerReq() {
	h.addReqRecord()
	reqQueue.Add(&h.trigger)
}

func (h *helper) checkTaskUpdateReq() error {
	if h.trigger.Name == "" {
		return fmt.Errorf("trigger id is required")
	}

	if h.trigger.Status == nil {
		return fmt.Errorf("trigger status is required")
	}

	return nil
}

func (h *helper) parseTriggerEnablement() error {
	err := h.c.ShouldBindJSON(&h.toggle)
	if err != nil {
		return err
	}

	name := h.c.Param("triggerName")
	trigger, found := trigger.Get(name)
	if !found {
		return fmt.Errorf("trigger(%s): trigger not found", h.trigger.Name)
	}

	h.trigger = *trigger
	h.trigger.Enabled = h.toggle.Enable
	h.trigger.InitUpdateStatus()
	return nil
}
