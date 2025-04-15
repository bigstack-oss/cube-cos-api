package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/gin-gonic/gin"
)

// func checkSenderUpdate(c *gin.Context) error {
// 	host := c.Param("senderHost")
// 	if !isSenderExist(host) {
// 		return errors.New("sender not found")
// 	}

// 	return nil
// }

func checkRecipientUpdate(c *gin.Context) error {
	recipientEmail := c.Param("recipientEmail")
	if !isRecipientExist(recipientEmail) {
		return errors.New("recipient not found")
	}

	err := email.CheckFormat(recipientEmail)
	if err != nil {
		return errors.New("recipient email format is invalid")
	}

	return nil
}

func checkSlackChannelUpdate(c *gin.Context) error {
	name := c.Param("channelName")
	if !isChannelExist(name) {
		return errors.New("channel not found")
	}

	return nil
}
