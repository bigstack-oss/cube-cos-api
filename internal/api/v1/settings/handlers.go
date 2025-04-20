package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
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
		api.SetBadRequest(c, err)
		return
	}

	if h.isSenderExist(h.task.Sender.Host) {
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
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get email recipients: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	sender := setting.Sender.Email.ConvertToApiSchema()
	api.SetStatusOk(
		c,
		"email senders retrieved successfully",
		[]email.Sender{sender},
	)
}

func patchEmailSender(c *gin.Context) {
	h, err := initReqHelper(c, "patchEmailSender")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !h.isSenderExist(h.task.Sender.Host) {
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

	if !h.isSenderExist(h.task.Sender.Host) {
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
	h, err := initReqHelper(c, "createEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if h.isRecipientExist() {
		api.SetStatusConflict(c, errors.New("recipient already exists"))
		return
	}

	h.updateSetting()
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
	policy, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get email recipients: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"email recipients retrieved successfully",
		policy.Receiver.Emails,
	)
}

func patchEmailRecipient(c *gin.Context) {
	h, err := initReqHelper(c, "patchEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = h.checkRecipientUpdate()
	if err != nil {
		log.Errorf("settings(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"email recipient updated successfully",
	)
}

func deleteEmailRecipient(c *gin.Context) {
	h, err := initReqHelper(c, "deleteEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !h.isRecipientExist() {
		api.SetBadRequest(c, errors.New("recipient not found"))
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"email recipient updated successfully",
	)
}

func createSlackChannel(c *gin.Context) {
	h, err := initReqHelper(c, "createSlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if h.isSlackChannlExist() {
		api.SetStatusConflict(c, errors.New("sender host already exists"))
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"slack channel creation requested successfully",
	)
}

func trySlackChannel(c *gin.Context) {
	channel, err := cubecos.GetSlackChannel(c.Param("channelName"))
	if err != nil {
		log.Errorf("settings(%s): failed to get slack channel: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = sendTrialSlackMessage(channel)
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
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get slack channels: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"slack channels retrieved successfully",
		convertToApiSlackChannels(setting.Receiver.Slacks),
	)
}

func putSlackChannel(c *gin.Context) {
	h, err := initReqHelper(c, "putSlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !h.isSlackChannlExist() {
		api.SetBadRequest(c, errors.New("channel not found"))
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"slack channel updated successfully",
	)
}

func deleteSlackChannel(c *gin.Context) {
	h, err := initReqHelper(c, "deleteSlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if !h.isSlackChannlExist() {
		api.SetBadRequest(c, errors.New("channel not found"))
		return
	}

	h.updateSetting()
	api.SetStatusAccepted(
		c,
		"slack channel deleted successfully",
	)
}

func updateSettingTask(c *gin.Context) {
	h, err := initReqHelper(c, "updateSettingTask")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %s", api.GetReqId(c), err.Error())
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
