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
		Method:  "PUT",
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
		Path:    "/settings/email/senders/:senderHost",
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
		Method:  "PUT",
		Path:    "/settings/email/senders/:senderHost",
		Func:    patchEmailSender,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/email/senders/:senderHost",
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
		Method:  "POST",
		Path:    "/settings/email/recipients/:recipientEmail",
		Func:    tryEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/email/recipients",
		Func:    listEmailRecipients,
	},
	{
		Version: api.V1,
		Method:  "PUT",
		Path:    "/settings/email/recipients/:recipientEmail",
		Func:    patchEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/email/recipients/:recipientEmail",
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
		Path:    "/settings/slack/channels/:channelName",
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
		Method:  "PUT",
		Path:    "/settings/slack/channels/:channelName",
		Func:    putSlackChannel,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/slack/channels/:channelName",
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

	if isSenderExist(sender.Host) {
		api.SetStatusConflict(c, errors.New("sender host already exists"))
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

	err = recipient.CheckEmailFormat()
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
		api.SetBadRequest(c, errors.New("no email senders found"))
		return
	}

	sender := senders[0]
	err = sendTrialEmail(sender, recipient.Email)
	if err != nil {
		log.Errorf("request(%s): failed to try email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = setSenderAsVerified(sender)
	if err != nil {
		log.Errorf("request(%s): failed to enable email trial toggle: %s", api.GetReqId(c), err.Error())
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
	sender := email.Sender{}
	err := c.ShouldBindJSON(&sender)
	if err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
		return
	}

	err = checkSenderUpdate(c, sender)
	if err != nil {
		log.Errorf("request(%s): failed to check email sender: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
	}

	sender.Host = c.Param("senderHost")
	sender.ResetAccessVerification()
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
	host := c.Param("senderHost")
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

	err = recipient.CheckEmailFormat()
	if err != nil {
		log.Errorf("request(%s): invalid email format: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if isRecipientExist(recipient.Email) {
		api.SetStatusConflict(c, errors.New("recipient already exists"))
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

func tryEmailRecipient(c *gin.Context) {
	recipientEmail := c.Param("recipientEmail")
	if !isRecipientExist(recipientEmail) {
		api.SetBadRequest(c, errors.New("recipient not found"))
		return
	}

	senders, err := getEmailSenders()
	if err != nil {
		log.Errorf("request(%s): failed to get email senders: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}
	if len(senders) == 0 {
		api.SetBadRequest(c, errors.New("no email senders found"))
		return
	}

	sender := senders[0]
	if !sender.AccessVerified {
		api.SetBadRequest(c, errors.New("email sender not verified"))
		return
	}

	err = sendTrialEmail(sender, recipientEmail)
	if err != nil {
		log.Errorf("request(%s): failed to try email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email recipient tried successfully",
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
	recipient := email.Recipient{}
	err := c.ShouldBindJSON(&recipient)
	if err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	err = checkRecipientUpdate(c, recipient)
	if err != nil {
		log.Errorf("request(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	recipient.Email = c.Param("recipientEmail")
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
	recipient := c.Param("recipientEmail")
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

	if isChannelExist(channel.Name) {
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
	channel, err := getSlackChannel(c.Param("channelName"))
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
	channel := slack.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		log.Errorf("request(%s): failed to decode slack channel: %s", api.GetReqId(c), err.Error())
		return
	}

	err = checkSlackChannelUpdate(c, channel)
	if err != nil {
		log.Errorf("request(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	channel.Name = c.Param("channelName")
	err = updateSlackChannel(channel)
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
	name := c.Param("channelName")
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
