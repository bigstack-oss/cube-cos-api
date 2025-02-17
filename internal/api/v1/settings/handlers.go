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
