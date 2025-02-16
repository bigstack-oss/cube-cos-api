package settings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var Handlers = []api.Handler{
	{
		Version: api.V1,
		Method:  "POST",
		Path:    "/settings/email-sender",
		Func:    createEmailSender,
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
