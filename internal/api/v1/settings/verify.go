package settings

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/gin-gonic/gin"
)

func checkSenderUpdate(c *gin.Context, sender email.Sender) error {
	host := c.Param("senderHost")
	if !isSenderExist(host) {
		return errors.New("sender not found")
	}

	if host != sender.Host {
		return errors.New("sender host does not match")
	}

	return nil
}

func checkRecipientUpdate(c *gin.Context, recipient email.Recipient) error {
	recipientEmail := c.Param("recipientEmail")
	if !isRecipientExist(recipientEmail) {
		return errors.New("recipient not found")
	}

	if recipientEmail != recipient.Email {
		return errors.New("recipient email does not match")
	}

	err := recipient.CheckFormat()
	if err != nil {
		return errors.New("recipient email format is invalid")
	}

	return nil
}

func checkSlackChannelUpdate(c *gin.Context, channel slack.Channel) error {
	name := c.Param("channelName")
	if !isChannelExist(name) {
		return errors.New("channel not found")
	}

	if name != channel.Name {
		return errors.New("channel name does not match")
	}

	return nil
}
