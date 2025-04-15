package email

import (
	"fmt"
	"net/mail"
	"net/smtp"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

const (
	SenderCollection    = "emailSenders"
	RecipientCollection = "emailRecipients"
)

type Options struct {
	Recipients []Recipient `json:"recipients" bson:"recipients"`
	Senders    []Sender    `json:"senders" bson:"senders"`
}

type Sender struct {
	Host           string           `json:"host,omitempty" bson:"host" yaml:"host,omitempty"`
	Port           int              `json:"port,omitempty" bson:"port" yaml:"port,omitempty"`
	Username       string           `json:"username,omitempty" bson:"username" yaml:"username,omitempty"`
	Password       string           `json:"password,omitzero" bson:"password" yaml:"password,omitempty"`
	Email          string           `json:"email,omitempty" bson:"email" yaml:"email,omitempty"`
	AccessVerified bool             `json:"accessVerified" bson:"accessVerified" yaml:"-"`
	Status         *status.Settings `json:"status,omitempty" bson:"status" yaml:"-"`
}

func (s *Sender) RequirePasswordChange() bool {
	return s.Password != ""
}

func (s *Sender) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *Sender) UserAuth() smtp.Auth {
	return smtp.PlainAuth("", s.Username, s.Password, s.Host)
}

func (s *Sender) ResetAccessVerification() {
	s.AccessVerified = false
}

func (s *Sender) ErasePassword() {
	s.Password = ""
}

func (s *Sender) InitOkStatus() {
	s.Status = &status.Settings{
		Current:    status.Ok,
		IsUpdating: false,
	}
}

func (s *Sender) InitUpdateStatus() {
	s.Status = &status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}

type Recipient struct {
	Address string          `json:"address" bson:"address"`
	Note    string          `json:"note" bson:"note"`
	Enabled bool            `json:"enabled,omitempty" bson:"-"`
	Status  status.Settings `json:"status" bson:"status" yaml:"-"`
}

type Trial struct {
	Email string `json:"email" bson:"email"`
}

func CheckFormat(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func (r *Recipient) InitUpdateStatus() {
	r.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}

func (o *Options) InitOkStatus() {
	for i := range o.Recipients {
		o.Recipients[i].Status.InitOkStatus()
	}

	for i := range o.Senders {
		o.Senders[i].InitOkStatus()
	}
}
