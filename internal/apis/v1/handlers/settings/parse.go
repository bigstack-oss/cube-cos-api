package settings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	log "go-micro.dev/v5/logger"
)

func (h *helper) initTitlePrefixUpdateParams() error {
	h.task = &setting.Options{Type: "titlePrefix"}
	err := h.c.ShouldBindJSON(&h.task.TitlePrefix)
	if err != nil {
		return err
	}

	h.task.Key = h.task.TitlePrefix.Value
	h.task.Value = h.task.TitlePrefix.Value
	h.task.InitUpdateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initEmailSenderCreateParams() error {
	h.task = &setting.Options{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email sender: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	if h.task.Sender.IsHostEmpty() {
		log.Errorf("settings(%s): %v", queries.GetReqId(h.c), errors.ErrEmailSenderHostInvalid)
		return errors.ErrEmailSenderHostInvalid
	}

	if h.task.Sender.IsPortEmpty() {
		log.Errorf("settings(%s): %v", queries.GetReqId(h.c), errors.ErrEmailSenderPortInvalid)
		return errors.ErrEmailSenderPortInvalid
	}

	h.task.Key = h.task.Sender.Host
	h.task.InitUpdateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initEmailSenderPatchParams() error {
	h.emailSender = h.c.Param("senderHost")
	if h.emailSender == "" {
		return errors.ErrEmailSenderHostIsEmpty
	}

	h.task = &setting.Options{Type: "emailSender"}
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
	h.task.InitUpdateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

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

	h.task = &setting.Options{Type: "emailSender", Key: host}
	h.task.Sender = &email.Sender{Host: host}
	h.task.InitDeleteStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initEmailRecipientCreateParams() error {
	h.task = &setting.Options{Type: "emailRecipient"}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email recipient: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	err = email.CheckFormat(h.task.Recipient.Address)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %s", queries.GetReqId(h.c), err.Error())
		bodies.SetBadRequest(h.c, err)
		return err
	}

	h.task.Key = h.task.Recipient.Address
	h.task.InitCreateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initEmailRecipientTrialParams() error {
	h.recipientEmail = h.c.Param("recipientEmail")
	if !isRecipientExist(h.recipientEmail) {
		return errors.ErrEmailRecipientNotFound
	}

	return nil
}

func (h *helper) initEmailRecipientPatchParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.ErrEmailRecipientIsEmpty
	}

	h.task = &setting.Options{Type: "emailRecipient", Key: recipientEmail}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		return err
	}

	if h.task.Recipient.Address == "" {
		h.task.Recipient.Address = recipientEmail
	}

	h.task.InitUpdateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initEmailRecipientDeleteParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.ErrEmailRecipientIsEmpty
	}

	h.task = &setting.Options{Type: "emailRecipient", Key: recipientEmail}
	h.task.Recipient = &email.Recipient{Address: recipientEmail}
	h.task.InitDeleteStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initSlackChannelDeleteParams() error {
	channelName := h.c.Param("channelName")
	if channelName == "" {
		return errors.ErrSlackChannelNameIsEmpty
	}

	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get alert setting: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	h.task = &setting.Options{Type: "slackChannel", Key: channelName}
	h.task.Slack = &slack.ApiChannel{Name: channelName, URL: policy.GetSlackUrlByName(channelName)}
	h.task.InitDeleteStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initSlackChannelCreateParams() error {
	h.task = &setting.Options{Type: "slackChannel"}
	err := h.c.ShouldBindJSON(&h.task.Slack)
	if err != nil {
		log.Errorf("settings(%s): failed to decode slack channel: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	h.task.Key = h.task.Slack.Name
	h.task.InitCreateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) initSlackChannelPatchParams() error {
	channelName := h.c.Param("channelName")
	if channelName == "" {
		return errors.ErrSlackChannelNameIsEmpty
	}

	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get alert setting: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	h.task = &setting.Options{Type: "slackChannel", Key: policy.GetSlackUrlByName(channelName)}
	err = h.c.ShouldBindJSON(&h.task.Slack)
	if err != nil {
		return err
	}

	if h.task.Slack.Name == "" {
		h.task.Slack.Name = channelName
	}

	h.task.InitUpdateStatus()
	h.task.ShouldReportToController = h.isClusterWiseRequired

	return nil
}

func (h *helper) parseTaskUpdate() error {
	err := h.c.ShouldBindJSON(&h.task)
	if err != nil {
		log.Errorf("settings(%s): failed to parse task: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	if h.task.Type == "" {
		log.Errorf("settings(%s): %v", queries.GetReqId(h.c), errors.ErrAlertSettingTaskTypeIsEmpty)
		return err
	}

	return nil
}
