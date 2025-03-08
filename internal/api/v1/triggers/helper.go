package triggers

import (
	"fmt"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c *gin.Context
}

func initReqHelper(c *gin.Context) (*helper, error) {
	return &helper{c: c}, nil
}

func (h *helper) listTriggers() ([]trigger.Options, error) {
	triggers := []trigger.Options{}
	for _, trigger := range trigger.DefaultOptions {
		setResponseItemsToTrigger(&trigger)
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

func (h *helper) getTrigger(name string) (*trigger.Options, error) {
	for _, trigger := range trigger.DefaultOptions {
		if trigger.Name == name {
			setResponseItemsToTrigger(&trigger)
			return &trigger, nil
		}
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func setResponseItemsToTrigger(trigger *trigger.Options) {
	trigger.InitResponse()

	setEmailRecipientsToTrigger(trigger)
	if trigger.HasEmailRecipients() {
		trigger.Response.Types = append(
			trigger.Response.Types,
			"email",
		)
	}

	setSlackChannelsToTrigger(trigger)
	if trigger.HasSlackChannels() {
		trigger.Response.Types = append(
			trigger.Response.Types,
			"slack",
		)
	}
}

func setEmailRecipientsToTrigger(trigger *trigger.Options) {
	recipients, err := definition.GetEmailRecipients()
	if err != nil {
		return
	}

	trigger.Emails = recipients
}

func setSlackChannelsToTrigger(trigger *trigger.Options) {
	channels, err := definition.GetSlackChannels()
	if err != nil {
		return
	}

	trigger.Slacks = channels
}
