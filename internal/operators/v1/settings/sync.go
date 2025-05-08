package settings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

func (o *Operator) operateReq(setting setting.Options) error {
	switch setting.Status.Desired {
	case status.Created:
		return o.createSetting(setting)
	case status.Updated:
		return o.updateSetting(setting)
	case status.Deleted:
		return o.deleteSetting(setting)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for setting(%s)",
		setting.Status.Desired,
		setting.Type,
	)
}

func (o *Operator) createSetting(setting setting.Options) error {
	switch setting.Type {
	case "emailRecipient":
		return cubecos.ApplyEmailRecipient(*setting.Recipient)
	case "slackChannel":
		return cubecos.ApplySlackChannel(setting.Slack.ConvertToCosSchema())
	}

	return errors.ErrUnknownSettingType
}

func (o *Operator) updateSetting(setting setting.Options) error {
	switch setting.Type {
	case "titlePrefix":
		return cubecos.ApplyTitlePrefix(setting.TitlePrefix.Value)
	case "emailSender":
		return cubecos.ApplyEmailSender(*setting.Sender)
	case "emailRecipient":
		return cubecos.DeleteAndCreateEmailRecipient(setting)
	case "slackChannel":
		return cubecos.DeleteAndCreateSlackChannel(setting)
	}

	return errors.ErrUnknownSettingType
}

func (o *Operator) deleteSetting(setting setting.Options) error {
	switch setting.Type {
	case "emailSender":
		return cubecos.DeleteEmailSender()
	case "emailRecipient":
		return cubecos.DeleteEmailRecipient(setting.Recipient.Address)
	case "slackChannel":
		return cubecos.DeleteSlackChannel(setting.Slack.URL)
	}

	return errors.ErrUnknownSettingType
}
