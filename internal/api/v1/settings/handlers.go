package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/settings"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = settings.ReqQueue
	Handlers = []api.Handler{
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
			Method:  "PATCH",
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
		{
			Version: api.V1,
			Method:  "PATCH",
			Path:    "/settings/tasks",
			Func:    updateSettingTask,
		},
	}
)

func listSettings(c *gin.Context) {
	h, err := initReqHelper(c, "listSettings")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	setting, err := h.listSettings()
	if err != nil {
		log.Errorf("settings(%s): failed to get setting: %s", api.GetReqId(c), err.Error())
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
	h, err := initReqHelper(c, "patchTitlePrefix")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"title prefix update request successfully",
	)
}

func createEmailSender(c *gin.Context) {
	h, err := initReqHelper(c, "createEmailSender")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if h.isSenderExist() {
		api.SetStatusConflict(c, errors.New("sender host already exists"))
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"email sender created successfully",
	)
}

func tryEmailSender(c *gin.Context) {
	trial := email.Trial{}
	err := c.ShouldBindJSON(&trial)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = email.CheckFormat(trial.Email)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	senders, err := cubecos.GetEmailSenders()
	if err != nil {
		log.Errorf("settings(%s): failed to get email senders: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}
	if len(senders) == 0 {
		api.SetBadRequest(c, errors.New("no email senders found"))
		return
	}

	sender := senders[0]
	err = sendTrialEmail(sender, trial.Email)
	if err != nil {
		log.Errorf("settings(%s): failed to try email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = setSenderAsVerified(sender)
	if err != nil {
		log.Errorf("settings(%s): failed to enable email trial toggle: %s", api.GetReqId(c), err.Error())
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
	senders, err := definition.GetEmailSenders()
	if err != nil {
		log.Errorf("settings(%s): failed to list email senders: %s", api.GetReqId(c), err.Error())
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
	h, err := initReqHelper(c, "patchEmailSender")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !h.isSenderExist() {
		api.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	h.resetAccessVerification()
	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"email sender update requested successfully",
	)
}

func deleteEmailSender(c *gin.Context) {
	h, err := initReqHelper(c, "deleteEmailSender")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !h.isSenderExist() {
		api.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"email sender deletion requested successfully",
	)
}

func createEmailRecipient(c *gin.Context) {
	recipient := email.Recipient{}
	err := c.ShouldBindJSON(&recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	err = email.CheckFormat(recipient.Address)
	if err != nil {
		log.Errorf("settings(%s): invalid email format: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if isRecipientExist(recipient.Address) {
		api.SetStatusConflict(c, errors.New("recipient already exists"))
		return
	}

	err = insertEmailRecipient(recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to create email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusAccepted(
		c,
		"email recipient created successfully",
	)
}

func tryEmailRecipient(c *gin.Context) {
	recipientEmail := c.Param("recipientEmail")
	if !isRecipientExist(recipientEmail) {
		api.SetBadRequest(c, errors.New("recipient not found"))
		return
	}

	senders, err := definition.GetEmailSenders()
	if err != nil {
		log.Errorf("settings(%s): failed to get email senders: %s", api.GetReqId(c), err.Error())
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
		log.Errorf("settings(%s): failed to try email recipient: %s", api.GetReqId(c), err.Error())
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
	recipients, err := definition.GetEmailRecipients()
	if err != nil {
		log.Errorf("settings(%s): failed to list email recipients: %s", api.GetReqId(c), err.Error())
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
		log.Errorf("settings(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	err = checkRecipientUpdate(c)
	if err != nil {
		log.Errorf("settings(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = parseEmailRecipientUpdate(c, &recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to parse email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = updateEmailRecipient(c, recipient)
	if err != nil {
		log.Errorf("settings(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
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
		log.Errorf("settings(%s): failed to delete email recipient: %s", api.GetReqId(c), err.Error())
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
		log.Errorf("settings(%s): failed to decode slack channel: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if isChannelExist(channel.Name) {
		api.SetBadRequest(c, errors.New("channel already exists"))
		return
	}

	err = insertSlackChannel(channel)
	if err != nil {
		log.Errorf("settings(%s): failed to create slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusAccepted(
		c,
		"slack channel created successfully",
	)
}

func trySlackChannel(c *gin.Context) {
	channel, err := getSlackChannel(c.Param("channelName"))
	if err != nil {
		log.Errorf("settings(%s): failed to get slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = sendTrialSlackMessage(*channel)
	if err != nil {
		log.Errorf("settings(%s): failed to try slack channel: %s", api.GetReqId(c), err.Error())
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
	channels, err := definition.GetSlackChannels()
	if err != nil {
		log.Errorf("settings(%s): failed to list slack channels: %s", api.GetReqId(c), err.Error())
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
	err := checkSlackChannelUpdate(c)
	if err != nil {
		log.Errorf("settings(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	channel, err := parseSlackChannelUpdate(c)
	if err != nil {
		log.Errorf("settings(%s): failed to parse slack channel: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = updateSlackChannel(c, *channel)
	if err != nil {
		log.Errorf("settings(%s): failed to update slack channel: %s", api.GetReqId(c), err.Error())
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
		log.Errorf("settings(%s): failed to delete slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"slack channel deleted successfully",
		nil,
	)
}

func updateSettingTask(c *gin.Context) {
	h, err := initReqHelper(c, "updateSettingTask")
	if err != nil {
		log.Errorf("tunings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.parseTaskUpdate()
	if err != nil {
		api.SetBadRequest(c, err)
		return
	}

	err = h.updateSettingTask()
	if err != nil {
		log.Errorf("settings(%s): failed to update setting task: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"setting status updated",
		nil,
	)
}
