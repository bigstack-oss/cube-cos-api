package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/settings"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = settings.ReqQueue
	Handlers = []apis.Handler{
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/settings",
			Func:    listSettings,
		},
		{
			Version: apis.V1,
			Method:  "PUT",
			Path:    "/settings/titlePrefix",
			Func:    updateTitlePrefix,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/settings/email/senders",
			Func:    createEmailSender,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/settings/email/senders/:senderHost",
			Func:    tryEmailSender,
		},
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/settings/email/senders",
			Func:    listEmailSenders,
		},
		{
			Version: apis.V1,
			Method:  "PATCH",
			Path:    "/settings/email/senders/:senderHost",
			Func:    patchEmailSender,
		},
		{
			Version: apis.V1,
			Method:  "DELETE",
			Path:    "/settings/email/senders/:senderHost",
			Func:    deleteEmailSender,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/settings/email/recipients",
			Func:    createEmailRecipient,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/settings/email/recipients/:recipientEmail",
			Func:    tryEmailRecipient,
		},
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/settings/email/recipients",
			Func:    listEmailRecipients,
		},
		{
			Version: apis.V1,
			Method:  "PUT",
			Path:    "/settings/email/recipients/:recipientEmail",
			Func:    patchEmailRecipient,
		},
		{
			Version: apis.V1,
			Method:  "DELETE",
			Path:    "/settings/email/recipients/:recipientEmail",
			Func:    deleteEmailRecipient,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/settings/slack/channels",
			Func:    createSlackChannel,
		},
		{
			Version: apis.V1,
			Method:  "POST",
			Path:    "/settings/slack/channels/:channelName",
			Func:    trySlackChannel,
		},
		{
			Version: apis.V1,
			Method:  "GET",
			Path:    "/settings/slack/channels",
			Func:    listSlackChannels,
		},
		{
			Version: apis.V1,
			Method:  "PUT",
			Path:    "/settings/slack/channels/:channelName",
			Func:    putSlackChannel,
		},
		{
			Version: apis.V1,
			Method:  "DELETE",
			Path:    "/settings/slack/channels/:channelName",
			Func:    deleteSlackChannel,
		},
		{
			Version: apis.V1,
			Method:  "PATCH",
			Path:    "/settings/tasks",
			Func:    updateSettingTask,
		},
	}
)

func listSettings(c *gin.Context) {
	h, err := initHelper(c, "listSettings")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	setting, err := h.listSettings()
	if err != nil {
		log.Errorf("settings(%s): failed to get setting: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"all setting retrieved successfully",
		setting,
	)
}

func updateTitlePrefix(c *gin.Context) {
	h, err := initHelper(c, "updateTitlePrefix")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"title prefix update request successfully",
	)
}

func createEmailSender(c *gin.Context) {
	h, err := initHelper(c, "createEmailSender")
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	if h.isSenderExist(h.task.Sender.Host) {
		bodies.SetConflict(c, errors.New("sender host already exists"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"email sender created successfully",
	)
}

func tryEmailSender(c *gin.Context) {
	h, err := initHelper(c, "tryEmailSender")
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	senders, err := cubecos.GetEmailSenders()
	if err != nil {
		log.Errorf("settings(%s): failed to get email senders: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}
	if len(senders) == 0 {
		bodies.SetBadRequest(c, errors.New("no email senders found"))
		return
	}

	sender := senders[0]
	err = h.sendEmail(&sender, h.trial.Email)
	if err != nil {
		log.Errorf("settings(%s): %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.setSenderAsVerified(sender)
	if err != nil {
		log.Errorf("settings(%s): failed to mark sender as verified: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"email sender tried successfully",
		nil,
	)
}

func listEmailSenders(c *gin.Context) {
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to list email recipients: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	apiSchema := setting.ToApiSchema()
	bodies.SetOk(
		c,
		"email senders retrieved successfully",
		apiSchema.Email.Senders,
	)
}

func patchEmailSender(c *gin.Context) {
	h, err := initHelper(c, "patchEmailSender")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if !h.isSenderExist(h.emailSender) {
		bodies.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	h.resetAccessVerification()
	h.updateEmailSenderRecord()
	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"email sender update requested successfully",
	)
}

func deleteEmailSender(c *gin.Context) {
	h, err := initHelper(c, "deleteEmailSender")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if !h.isSenderExist(h.task.Sender.Host) {
		bodies.SetBadRequest(c, errors.New("sender not found"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"email sender deletion requested successfully",
	)
}

func createEmailRecipient(c *gin.Context) {
	h, err := initHelper(c, "createEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if h.isRecipientExist(h.task.Recipient.Address) {
		bodies.SetConflict(c, errors.New("recipient already exists"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"email recipient created successfully",
	)
}

func tryEmailRecipient(c *gin.Context) {
	h, err := initHelper(c, "tryEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	sender, err := h.getVerifiedSender()
	if err != nil {
		log.Errorf("settings(%s): failed to get verified email sender: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	err = h.sendEmail(sender, h.recipientEmail)
	if err != nil {
		log.Errorf("settings(%s): failed to try email recipient: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"email recipient tried successfully",
		nil,
	)
}

func listEmailRecipients(c *gin.Context) {
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get email recipients: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"email recipients retrieved successfully",
		setting.Receiver.Emails,
	)
}

func patchEmailRecipient(c *gin.Context) {
	h, err := initHelper(c, "patchEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.checkRecipientUpdate()
	if err != nil {
		log.Errorf("settings(%s): failed to update email recipient: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"email recipient updated successfully",
	)
}

func deleteEmailRecipient(c *gin.Context) {
	h, err := initHelper(c, "deleteEmailRecipient")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if !h.isRecipientExist(h.c.Param("recipientEmail")) {
		bodies.SetBadRequest(c, errors.New("recipient not found"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"email recipient updated successfully",
	)
}

func createSlackChannel(c *gin.Context) {
	h, err := initHelper(c, "createSlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if h.isSlackChannlExist() {
		bodies.SetConflict(c, errors.New("slack channel already exists"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"slack channel creation requested successfully",
	)
}

func trySlackChannel(c *gin.Context) {
	h, err := initHelper(c, "trySlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.sendSlackMessage()
	if err != nil {
		log.Errorf("settings(%s): failed to try slack channel: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"slack channel tried successfully",
		nil,
	)
}

func listSlackChannels(c *gin.Context) {
	setting, err := cubecos.GetAlertSetting()
	if err != nil {
		log.Errorf("settings(%s): failed to get slack channels: %v", queries.GetReqId(c), err)
		bodies.SetInternalServerError(c, err)
		return
	}

	apiSchema := setting.ToApiSchema()
	bodies.SetOk(
		c,
		"slack channels retrieved successfully",
		apiSchema.Slack.Channels,
	)
}

func putSlackChannel(c *gin.Context) {
	h, err := initHelper(c, "putSlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if !h.isSlackChannlExist() {
		bodies.SetBadRequest(c, errors.New("channel not found"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"slack channel updated successfully",
	)
}

func deleteSlackChannel(c *gin.Context) {
	h, err := initHelper(c, "deleteSlackChannel")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	if !h.isSlackChannlExist() {
		bodies.SetBadRequest(c, errors.New("channel not found"))
		return
	}

	h.updateToAllControllers()
	bodies.SetAccepted(
		c,
		"slack channel deleted successfully",
	)
}

func updateSettingTask(c *gin.Context) {
	h, err := initHelper(c, "updateSettingTask")
	if err != nil {
		log.Errorf("settings(%s): failed to init request helper: %v", h.reqId, err)
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.parseTaskUpdate()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.updateSettingTask()
	if err != nil {
		log.Errorf("settings(%s): failed to update setting task: %v", h.reqId, err)
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"setting status updated",
		nil,
	)
}
