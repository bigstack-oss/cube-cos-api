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
	return trigger.DefaultOptions, nil
}

func (h *helper) getTrigger(name string) (*trigger.Options, error) {
	for _, trigger := range trigger.DefaultOptions {
		if trigger.Name != name {
			continue
		}

		setEmailRecipientsToTrigger(&trigger)
		setSlackRecipientsToTrigger(&trigger)
		return &trigger, nil
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func setEmailRecipientsToTrigger(trigger *trigger.Options) {
	var err error
	trigger.Emails, err = definition.GetEmailRecipients()
	if err != nil {
		return
	}
}

func setSlackRecipientsToTrigger(trigger *trigger.Options) {
	var err error
	trigger.Slacks, err = definition.GetSlackChannels()
	if err != nil {
		return
	}
}
