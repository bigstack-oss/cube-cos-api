package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

var Handlers = []api.Handler{
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/email-sender",
		Func:    createEmailSender,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/email-sender",
		Func:    getEmailSender,
	},
	{
		Version: api.V1,
		Method:  "PUT",
		Path:    "/settings/email-sender",
		Func:    updateEmailSender,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/email-sender",
		Func:    deleteEmailSender,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/email-recipients",
		Func:    createEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/email-recipients",
		Func:    getEmailRecipients,
	},
	{
		Version: api.V1,
		Method:  "PUT",
		Path:    "/settings/email-recipients/:id",
		Func:    updateEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "DELETE",
		Path:    "/settings/email-recipients/:id",
		Func:    deleteEmailRecipient,
	},
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/slack-webhooks",
		Func:    createSlackWebhook,
	},
	{
		Version: api.V1,
		Method:  "GET",
		Path:    "/settings/slack-webhooks",
		Func:    getSlackWebhooks,
	},
	{
		Version: api.V1,
		Method:  "PUT",
		Path:    "/settings/slack-webhooks/:id",
		Func:    updateSlackWebhook,
	},
}

func createEmailSender(c *gin.Context) {
	var emailSender v1.EmailSender
	if err := c.ShouldBindJSON(&emailSender); err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
		return
	}

	if err := upsertEmailSenderRecord(emailSender); err != nil {
		log.Errorf("request(%s): failed to create email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusCreated(c, "email sender created successfully", nil)
}

func getEmailSender(c *gin.Context) {
	emailSender, err := getEmailSenderRecord()
	if errors.Is(err, mongo.ErrNoDocuments) {
		api.SetStatusNotFound(c, errors.New("email sender not found"))
		return
	}
	if err != nil {
		log.Errorf("request(%s): failed to get email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "email sender retrieved successfully", emailSender)
}

func updateEmailSender(c *gin.Context) {
	var emailSender v1.EmailSender
	if err := c.ShouldBindJSON(&emailSender); err != nil {
		log.Errorf("request(%s): failed to decode email sender: %s", api.GetReqId(c), err.Error())
		return
	}

	if err := upsertEmailSenderRecord(emailSender); err != nil {
		log.Errorf("request(%s): failed to update email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "email sender updated successfully", nil)
}

func deleteEmailSender(c *gin.Context) {
	if err := deleteEmailSenderRecord(); err != nil {
		log.Errorf("request(%s): failed to delete email sender: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "email sender deleted successfully", nil)
}

func createEmailRecipient(c *gin.Context) {
	var emailRecipient v1.EmailRecipient
	if err := c.ShouldBindJSON(&emailRecipient); err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	if err := createEmailRecipientRecord(emailRecipient); err != nil {
		log.Errorf("request(%s): failed to create email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusCreated(c, "email recipient created successfully", nil)
}

func getEmailRecipients(c *gin.Context) {
	emailRecipients, err := getEmailRecipientRecords()
	if err != nil {
		log.Errorf("request(%s): failed to get email recipients: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "email recipients retrieved successfully", emailRecipients)
}

func updateEmailRecipient(c *gin.Context) {
	var emailRecipient v1.EmailRecipient
	if err := c.ShouldBindJSON(&emailRecipient); err != nil {
		log.Errorf("request(%s): failed to decode email recipient: %s", api.GetReqId(c), err.Error())
		return
	}

	emailRecipient.ID = c.Param("id")
	if err := updateEmailRecipientRecord(emailRecipient); err != nil {
		log.Errorf("request(%s): failed to update email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "email recipient updated successfully", nil)
}

func deleteEmailRecipient(c *gin.Context) {
	recipientID := c.Param("id")
	if err := deleteEmailRecipientRecord(recipientID); err != nil {
		log.Errorf("request(%s): failed to delete email recipient: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "email recipient deleted successfully", nil)
}

func createSlackWebhook(c *gin.Context) {
	var slackWebhook v1.SlackWebhook
	if err := c.ShouldBindJSON(&slackWebhook); err != nil {
		log.Errorf("request(%s): failed to decode slack webhook: %s", api.GetReqId(c), err.Error())
		return
	}

	if err := createSlackWebhookRecord(slackWebhook); err != nil {
		log.Errorf("request(%s): failed to create slack webhook: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusCreated(c, "slack webhook created successfully", nil)
}

func getSlackWebhooks(c *gin.Context) {
	slackWebhooks, err := getSlackWebhookRecords()
	if err != nil {
		log.Errorf("request(%s): failed to get slack webhooks: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "slack webhooks retrieved successfully", slackWebhooks)
}

func updateSlackWebhook(c *gin.Context) {
	var slackWebhook v1.SlackWebhook
	if err := c.ShouldBindJSON(&slackWebhook); err != nil {
		log.Errorf("request(%s): failed to decode slack webhook: %s", api.GetReqId(c), err.Error())
		return
	}

	slackWebhook.ID = c.Param("id")
	if err := updateSlackWebhookRecord(slackWebhook); err != nil {
		log.Errorf("request(%s): failed to update slack webhook: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "slack webhook updated successfully", nil)
}
