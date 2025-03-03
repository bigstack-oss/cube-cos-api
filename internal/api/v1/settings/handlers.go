package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var Handlers = []api.Handler{
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings",
		Func:    listSettings,
	},
	{
		Version: api.V1,
		Method:  "PATCH",
		Path:    "/settings/titlePrefix",
		Func:    patchTitlePrefix,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/email/senders",
		Func:    createEmailSender,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/email/senders/:host",
		Func:    tryEmailSender,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/email/senders",
		Func:    listEmailSenders,
	},
	{
		Version: api.V1,
		Method:  "PATCH",
		Path:    "/settings/email/senders/:host",
		Func:    patchEmailSender,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/email/senders:host",
		Func:    deleteEmailSender,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/email/recipients",
		Func:    createEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/email/recipients",
		Func:    listEmailRecipients,
	},
	{
		Version: api.V1,
		Method:  "PATCH",
		Path:    "/settings/email/recipients/:email",
		Func:    patchEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/email/recipients/:email",
		Func:    deleteEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/slack/channels",
		Func:    createSlackChannel,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/slack/channels/:name",
		Func:    trySlackChannel,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/slack/channels",
		Func:    listSlackChannels,
	},
	{
		Version: api.V1,
		Method:  "PATCH",
		Path:    "/settings/slack/channels/:name",
		Func:    putSlackChannel,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/slack/channels/:name",
		Func:    deleteSlackChannel,
	},
}

func listSettings(c *gin.Context) {
	setting, err := getAllSettings()
	if err != nil {
		log.Errorf("request(%s): failed to get setting: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"all setting retrieved successfully",
		setting,
	)
}

func patchTitlePrefix(c *gin.Context) {
	titlePrefix := v1.TitlePrefix{}
	err := c.ShouldBindJSON(&titlePrefix)
	if err != nil {
		log.Errorf("request(%s): failed to decode title prefix: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = upsertTitlePrefix(titlePrefix.Value)
	if err != nil {
		log.Errorf("request(%s): failed to update title prefix: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"title prefix updated successfully",
		nil,
	)
}

func createEmailSender(c *gin.Context) {
	sender := email.Sender{}
	err := c.ShouldBindJSON(&sender)
	if err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !isSenderExist(sender.Host) {
		api.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	err = insertEmailSender(sender)
	if err != nil {
		log.Errorf("request(%s): failed to create email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusCreated(
		c,
		"email sender created successfully",
		nil,
	)
}

func tryEmailSender(c *gin.Context) {
	recipient := email.Recipient{}
	err := c.ShouldBindJSON(&recipient)
	if err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = recipient.CheckFormat()
	if err != nil {
		log.Errorf("request(%s): invalid email format: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	senders, err := getEmailSenders()
	if err != nil {
		log.Errorf("request(%s): failed to get email senders: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}
	if len(senders) == 0 {
		log.Error("no email senders found")
		api.SetBadRequest(c, errors.New("no email senders found"))
		return
	}

	err = sendTrialEmail(senders[0], recipient.Email)
	if err != nil {
		log.Errorf("request(%s): failed to try email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email sender tried successfully",
		nil,
	)
}

func listEmailSenders(c *gin.Context) {
	senders, err := getEmailSenders()
	if err != nil {
		log.Errorf("request(%s): failed to list email senders: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email senders retrieved successfully",
		senders,
	)
}

func patchEmailSender(c *gin.Context) {
	host := c.Param("host")
	if !isSenderExist(host) {
		api.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	sender := email.Sender{}
	err := c.ShouldBindJSON(&sender)
	if err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
		return
	}

	err = updateEmailSender(sender)
	if err != nil {
		log.Errorf("request(%s): failed to update email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email sender updated successfully",
		nil,
	)
}

func deleteEmailSender(c *gin.Context) {
	host := c.Param("host")
	if !isSenderExist(host) {
		api.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	err := removeEmailSender(host)
	if err != nil {
		log.Errorf("request(%s): failed to delete email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email sender deleted successfully",
		nil,
	)
}

func createEmailRecipient(c *gin.Context) {
	recipient := email.Recipient{}
	err := c.ShouldBindJSON(&recipient)
	if err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	err = recipient.CheckFormat()
	if err != nil {
		log.Errorf("request(%s): invalid email format: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !isRecipientExist(recipient.Email) {
		api.SetBadRequest(c, errors.New("recipient already exists"))
		return
	}

	err = insertEmailRecipient(recipient)
	if err != nil {
		log.Errorf("request(%s): failed to create email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusCreated(
		c,
		"email recipient created successfully",
		nil,
	)
}

func listEmailRecipients(c *gin.Context) {
	recipients, err := getEmailRecipients()
	if err != nil {
		log.Errorf("request(%s): failed to list email recipients: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email recipients retrieved successfully",
		recipients,
	)
}

func patchEmailRecipient(c *gin.Context) {
	recipientEmail := c.Param("email")
	if !isRecipientExist(recipientEmail) {
		api.SetBadRequest(c, errors.New("recipient not found"))
		return
	}

	recipient := email.Recipient{}
	err := c.ShouldBindJSON(&recipient)
	if err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	recipient.Email = recipientEmail
	err = updateEmailRecipient(recipient)
	if err != nil {
		log.Errorf("request(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email recipient updated successfully",
		nil,
	)
}

func deleteEmailRecipient(c *gin.Context) {
	recipient := c.Param("email")
	if !isRecipientExist(recipient) {
		api.SetBadRequest(c, errors.New("recipient not found"))
		return
	}

	err := removeEmailRecipient(recipient)
	if err != nil {
		log.Errorf("request(%s): failed to delete email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email recipient deleted successfully",
		nil,
	)
}

func createSlackChannel(c *gin.Context) {
	channel := slack.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		log.Errorf("request(%s): failed to decode slack channel: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !isChannelExist(channel.Name) {
		api.SetBadRequest(c, errors.New("channel already exists"))
		return
	}

	err = insertSlackChannel(channel)
	if err != nil {
		log.Errorf("request(%s): failed to create slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusCreated(
		c,
		"slack channel created successfully",
		nil,
	)
}

func trySlackChannel(c *gin.Context) {
	channel, err := getSlackChannel(c.Param("name"))
	if err != nil {
		log.Errorf("request(%s): failed to get slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = sendTrialSlackMessage(*channel)
	if err != nil {
		log.Errorf("request(%s): failed to try slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"slack channel tried successfully",
		nil,
	)
}

func listSlackChannels(c *gin.Context) {
	channels, err := getSlackChannels()
	if err != nil {
		log.Errorf("request(%s): failed to list slack channels: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"slack channels retrieved successfully",
		channels,
	)
}

func putSlackChannel(c *gin.Context) {
	name := c.Param("name")
	if !isChannelExist(name) {
		api.SetBadRequest(c, errors.New("channel not found"))
		return
	}

	channel := slack.Channel{}
	if err := c.ShouldBindJSON(&channel); err != nil {
		log.Errorf("request(%s): failed to decode slack channel: %s", api.GetReqId(c), err.Error())
		return
	}

	channel.Name = name
	err := updateSlackChannel(channel)
	if err != nil {
		log.Errorf("request(%s): failed to update slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"slack channel updated successfully",
		nil,
	)
}

func deleteSlackChannel(c *gin.Context) {
	name := c.Param("name")
	if !isChannelExist(name) {
		api.SetBadRequest(c, errors.New("channel not found"))
		return
	}

	err := removeSlackChannel(name)
	if err != nil {
		log.Errorf("request(%s): failed to delete slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"slack channel deleted successfully",
		nil,
	)
}
