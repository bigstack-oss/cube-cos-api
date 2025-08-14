package settings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "updateTitlePrefix":
		return h.initTitlePrefixUpdateParams()
	case "createEmailSender":
		return h.initEmailSenderCreateParams()
	case "tryEmailSender":
		return h.initEmailSenderTrialParams()
	case "patchEmailSender":
		return h.initEmailSenderPatchParams()
	case "deleteEmailSender":
		return h.initEmailSenderDeleteParams()
	case "createEmailRecipient":
		return h.initEmailRecipientCreateParams()
	case "tryEmailRecipient":
		return h.initEmailRecipientTrialParams()
	case "patchEmailRecipient":
		return h.initEmailRecipientPatchParams()
	case "deleteEmailRecipient":
		return h.initEmailRecipientDeleteParams()
	case "createSlackChannel":
		return h.initSlackChannelCreateParams()
	case "trySlackChannel":
		return h.initSlackChannelTrialParams()
	case "putSlackChannel":
		return h.initSlackChannelPatchParams()
	case "deleteSlackChannel":
		return h.initSlackChannelDeleteParams()
	case "updateSettingTask":
		return h.parseTaskUpdate()
	default:
		return nil
	}
}

func (h *helper) initTitlePrefixUpdateParams() error {
	h.task = &settings.Setting{Type: "titlePrefix"}
	err := h.c.ShouldBindJSON(&h.task.TitlePrefix)
	if err != nil {
		return err
	}

	h.task.Key = h.task.TitlePrefix.Value
	h.task.Value = h.task.TitlePrefix.Value
	h.task.SetUpdating()
	return nil
}

func (h *helper) initEmailSenderCreateParams() error {
	h.task = &settings.Setting{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email sender: %v", h.reqId, err)
		return err
	}

	if h.task.Sender.IsHostEmpty() {
		log.Errorf("settings(%s): %v", h.reqId, errors.ErrEmailSenderHostInvalid)
		return errors.ErrEmailSenderHostInvalid
	}

	if h.task.Sender.IsPortEmpty() {
		log.Errorf("settings(%s): %v", h.reqId, errors.ErrEmailSenderPortInvalid)
		return errors.ErrEmailSenderPortInvalid
	}

	h.task.Key = h.task.Sender.Host
	h.task.SetUpdating()
	return nil
}

func (h *helper) initEmailSenderTrialParams() error {
	err := h.c.ShouldBindJSON(&h.trial)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email: %v", h.reqId, err)
		return err
	}

	err = email.CheckFormat(h.trial.Email)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) initEmailSenderPatchParams() error {
	h.emailSender = h.c.Param("senderHost")
	if h.emailSender == "" {
		return errors.ErrEmailSenderHostIsEmpty
	}

	h.task = &settings.Setting{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		return err
	}

	if h.task.Sender.Host == "" {
		h.task.Sender.Host = h.emailSender
	}

	h.task.Key = h.task.Sender.Host
	h.task.Sender.Password = h.parsePassword()
	h.task.Sender.ResetAccessVerification()
	h.task.SetUpdating()
	return nil
}

func (h *helper) parsePassword() *string {
	if h.task.Sender.Password != nil {
		return h.task.Sender.Password
	}

	senders, err := cubecos.GetEmailSenders()
	if err != nil {
		return nil
	}
	if len(senders) == 0 {
		return nil
	}

	return senders[0].Password
}

func (h *helper) initEmailSenderDeleteParams() error {
	host := h.c.Param("senderHost")
	if host == "" {
		return errors.ErrEmailSenderHostIsEmpty
	}

	h.task = &settings.Setting{Type: "emailSender", Key: host}
	h.task.Sender = &email.Sender{Host: host}
	h.task.SetDeleting()
	return nil
}

func (h *helper) initEmailRecipientCreateParams() error {
	h.task = &settings.Setting{Type: "emailRecipient"}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email recipient: %v", h.reqId, err)
		return err
	}

	err = email.CheckFormat(h.task.Recipient.Address)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %v", h.reqId, err)
		bodies.SetBadRequest(h.c, err, nil)
		return err
	}

	h.task.Key = h.task.Recipient.Address
	h.task.SetCreating()
	return nil
}

func (h *helper) initEmailRecipientTrialParams() error {
	h.recipientEmail = h.c.Param("recipientEmail")
	if !h.isRecipientExist(h.recipientEmail) {
		return errors.ErrEmailRecipientNotFound
	}

	return nil
}

func (h *helper) initEmailRecipientPatchParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.ErrEmailRecipientIsEmpty
	}

	h.task = &settings.Setting{Type: "emailRecipient", Key: recipientEmail}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		return err
	}

	h.task.SetUpdating()
	if h.task.Recipient.Address == "" {
		h.task.Recipient.Address = recipientEmail
	}

	return nil
}

func (h *helper) initEmailRecipientDeleteParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.ErrEmailRecipientIsEmpty
	}

	h.task = &settings.Setting{Type: "emailRecipient", Key: recipientEmail}
	h.task.Recipient = &email.Recipient{Address: recipientEmail}
	h.task.SetDeleting()
	return nil
}

func (h *helper) initSlackChannelDeleteParams() error {
	channel := h.c.Param("channelName")
	if channel == "" {
		return errors.ErrSlackChannelNameIsEmpty
	}

	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get alert setting: %v", h.reqId, err)
		return err
	}

	h.task = &settings.Setting{Type: "slackChannel", Key: channel}
	h.task.SetDeleting()
	h.task.Slack = &slack.ApiChannel{
		Name: channel,
		URL:  policy.GetSlackUrlByName(channel),
	}

	return nil
}

func (h *helper) initSlackChannelCreateParams() error {
	h.task = &settings.Setting{Type: "slackChannel"}
	err := h.c.ShouldBindJSON(&h.task.Slack)
	if err != nil {
		log.Errorf("settings(%s): failed to decode slack channel: %v", h.reqId, err)
		return err
	}

	h.task.Key = h.task.Slack.URL
	h.task.SetCreating()
	return nil
}

func (h *helper) initSlackChannelTrialParams() error {
	channel, err := cubecos.GetSlackChannel(h.c.Param("channelName"))
	if err != nil {
		return err
	}

	h.slackChannel = channel.URL
	return nil
}

func (h *helper) initSlackChannelPatchParams() error {
	channel := h.c.Param("channelName")
	if channel == "" {
		return errors.ErrSlackChannelNameIsEmpty
	}

	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		return err
	}

	h.task = &settings.Setting{Type: "slackChannel", Key: policy.GetSlackUrlByName(channel)}
	err = h.c.ShouldBindJSON(&h.task.Slack)
	if err != nil {
		return err
	}

	h.task.SetUpdating()
	if h.task.Slack.Name == "" {
		h.task.Slack.Name = channel
	}

	return nil
}

func (h *helper) parseTaskUpdate() error {
	err := h.c.ShouldBindJSON(&h.task)
	if err != nil {
		log.Errorf("settings(%s): failed to parse task: %v", h.reqId, err)
		return err
	}

	if h.task.Type == "" {
		log.Errorf("settings(%s): %s", h.reqId, errors.ErrAlertSettingTaskTypeIsEmpty.Error())
		return err
	}

	h.hostname = h.c.Param("nodeName")
	if h.hostname == "" {
		err := fmt.Errorf("hostname cannot be empty for setting task update")
		log.Errorf("settings(%s): %v", h.reqId, err)
		return err
	}

	return nil
}
