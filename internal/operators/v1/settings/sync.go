package settings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

func (o *Operator) operateReq(setting setting.Options) error {
	switch setting.Status.Desired {
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

func (o *Operator) updateSetting(setting setting.Options) error {
	var err error

	switch setting.Type {
	case "titlePrefix":
		err = cubecos.ApplyTitlePrefix(setting.TitlePrefix.Value)
	case "emailSender":
		err = cubecos.ApplyEmailSender(*setting.Sender)
	case "emailRecipient":
		err = cubecos.ApplyEmailRecipient(*setting.Recipient)
	case "slackChannel":
		err = cubecos.ApplySlackChannel(setting.Slack.ConvertToCosSchema())
	default:
		return cuberr.UnknownSettingType
	}

	return err
}

func (o *Operator) deleteSetting(setting setting.Options) error {
	var err error

	switch setting.Type {
	case "titlePrefix":
		err = cubecos.ApplyTitlePrefix(setting.TitlePrefix.Value)
	case "emailSender":
		err = cubecos.DeleteEmailSender()
	case "emailRecipient":
		err = cubecos.DeleteEmailRecipient(setting.Recipient.Address)
	case "slackChannel":
		err = cubecos.DeleteSlackChannel(setting.Slack.URL)
	default:
		return cuberr.UnknownSettingType
	}

	return err
}
