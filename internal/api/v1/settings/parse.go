package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	log "go-micro.dev/v5/logger"
)

func (h *helper) initTitlePrefixPatchParams() error {
	h.task = &setting.Options{Type: "titlePrefix"}
	err := h.c.ShouldBindJSON(&h.task.TitlePrefix)
	if err != nil {
		return err
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailSenderCreateParams() error {
	h.task = &setting.Options{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email sender: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	if h.task.Sender.IsHostEmpty() {
		log.Errorf("settings(%s): %v", api.GetReqId(h.c), cuberr.EmailSenderHostInvalid)
		return cuberr.EmailSenderHostInvalid
	}

	if h.task.Sender.IsPortEmpty() {
		log.Errorf("settings(%s): %v", api.GetReqId(h.c), cuberr.EmailSenderPortInvalid)
		return cuberr.EmailSenderPortInvalid
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailSenderPatchParams() error {
	host := h.c.Param("senderHost")
	if host == "" {
		return errors.New("email sender host is empty")
	}

	h.task = &setting.Options{Type: "emailSender"}
	err := h.c.ShouldBindJSON(&h.task.Sender)
	if err != nil {
		return err
	}

	if h.task.Sender.Host == "" {
		h.task.Sender.Host = host
	}

	h.task.Sender.ResetAccessVerification()
	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailSenderDeleteParams() error {
	host := h.c.Param("senderHost")
	if host == "" {
		return errors.New("email sender host is empty")
	}

	h.task = &setting.Options{Type: "emailSender"}
	h.task.Sender = &email.Sender{Host: host}
	h.task.InitDeleteStatus()
	return nil
}

func (h *helper) initEmailRecipientCreateParams() error {
	h.task = &setting.Options{Type: "emailRecipient"}
	err := h.c.ShouldBindJSON(&h.task.Recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email recipient: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	err = email.CheckFormat(h.task.Recipient.Address)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %s", api.GetReqId(h.c), err.Error())
		api.SetBadRequest(h.c, err)
		return err
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initEmailRecipientPatchParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.New("email recipient email is empty")
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
	return nil
}

func (h *helper) initEmailRecipientDeleteParams() error {
	recipientEmail := h.c.Param("recipientEmail")
	if recipientEmail == "" {
		return errors.New("email recipient email is empty")
	}

	h.task = &setting.Options{Type: "emailRecipient"}
	h.task.Recipient = &email.Recipient{Address: recipientEmail}
	h.task.InitDeleteStatus()
	return nil
}

func (h *helper) initSlackChannelDeleteParams() error {
	channelName := h.c.Param("channelName")
	if channelName == "" {
		return errors.New("slack channel name is empty")
	}

	h.task = &setting.Options{Type: "slackChannel", Key: channelName}
	h.task.Slack = &slack.ApiChannel{Name: channelName}
	h.task.InitDeleteStatus()
	return nil
}

func (h *helper) initSlackChannelCreateParams() error {
	h.task = &setting.Options{Type: "slackChannel"}
	err := h.c.ShouldBindJSON(&h.task.Slack)
	if err != nil {
		log.Errorf("settings(%s): failed to decode slack channel: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	h.task.Key = h.task.Slack.Name
	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) initSlackChannelPatchParams() error {
	channelName := h.c.Param("channelName")
	if channelName == "" {
		return errors.New("slack channel name is empty")
	}

	h.task = &setting.Options{Type: "slackChannel", Key: channelName}
	err := h.c.ShouldBindJSON(&h.task.Slack)
	if err != nil {
		return err
	}

	if h.task.Slack.Name == "" {
		h.task.Slack.Name = channelName
	}

	h.task.InitUpdateStatus()
	return nil
}

func (h *helper) parseTaskUpdate() error {
	err := h.c.ShouldBindJSON(&h.task)
	if err != nil {
		log.Errorf("settings(%s): failed to parse task: %s", api.GetReqId(h.c), err.Error())
		return err
	}

	if h.task.Type == "" {
		err := errors.New("task type is empty")
		log.Errorf("settings(%s): %v", api.GetReqId(h.c), err)
		return err
	}

	if h.task.Key == "" {
		err := errors.New("task key is empty")
		log.Errorf("settings(%s): %v", api.GetReqId(h.c), err)
		return err
	}

	return nil
}
