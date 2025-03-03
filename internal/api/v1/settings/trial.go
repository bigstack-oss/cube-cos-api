package settings

import (
	"net/smtp"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
)

func sendTrialEmail(sender email.Sender, recipient string) error {
	return smtp.SendMail(
		sender.Address(),
		sender.UserAuth(),
		sender.Email,
		[]string{recipient},
		[]byte("Subject: Trial Email\n\nThis is a trial email from Cube COS."),
	)
}
