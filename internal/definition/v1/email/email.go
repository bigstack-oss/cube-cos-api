package email

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	SenderCollection    = "emailSenders"
	RecipientCollection = "emailRecipients"
)

type Options struct {
	Recipients []Recipient `json:"recipients" bson:"recipients"`
	Senders    []Sender    `json:"senders" bson:"senders"`
}

type CosSender struct {
	Host     string `json:"host,omitempty" bson:"host" yaml:"host,omitempty"`
	Port     string `json:"port,omitempty" bson:"port" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" bson:"username" yaml:"username,omitempty"`
	Password string `json:"password,omitzero" bson:"password" yaml:"password,omitempty"`
	From     string `json:"from,omitempty" bson:"from" yaml:"from,omitempty"`
}

func (c *CosSender) ToApiSchema() Sender {
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		port = 0
	}

	return Sender{
		Host:     c.Host,
		Port:     port,
		Username: c.Username,
		Password: &c.Password,
		From:     c.From,
	}
}

type Sender struct {
	Host           string           `json:"host,omitempty" bson:"host" yaml:"host,omitempty"`
	Port           int              `json:"port,omitempty" bson:"port" yaml:"port,omitempty"`
	Username       string           `json:"username,omitempty" bson:"username" yaml:"username,omitempty"`
	Password       *string          `json:"password,omitzero" bson:"password" yaml:"password,omitempty"`
	From           string           `json:"from,omitempty" bson:"from" yaml:"from,omitempty"`
	AccessVerified bool             `json:"accessVerified" bson:"accessVerified" yaml:"-"`
	Status         *status.Settings `json:"status,omitempty" bson:"status" yaml:"-"`
}

func (s *Sender) IsHostEmpty() bool {
	return s.Host == ""
}

func (s *Sender) IsPortEmpty() bool {
	return s.Port == 0
}

func (s *Sender) RequirePasswordChange() bool {
	return s.Password != nil
}

func (s *Sender) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *Sender) UserAuth() smtp.Auth {
	return smtp.PlainAuth("", s.Username, *s.Password, s.Host)
}

func (s *Sender) ResetAccessVerification() {
	s.AccessVerified = false
}

func (s *Sender) ErasePassword() {
	s.Password = nil
}

func (s *Sender) SetOk() {
	s.Status = &status.Settings{
		Current:    status.Ok,
		IsUpdating: false,
	}
}

func (s *Sender) SetUpdating() {
	s.Status = &status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}

type Recipient struct {
	Address string          `json:"address" bson:"address"`
	Note    string          `json:"note" bson:"note"`
	Enabled bool            `json:"enabled,omitempty" bson:"enabled"`
	Status  status.Settings `json:"status" bson:"status" yaml:"-"`
}

type Trial struct {
	Email string `json:"email" bson:"email"`
}

func CheckFormat(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func (r *Recipient) SetUpdating() {
	r.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
	}
}

func (o *Options) SetOk() {
	for i := range o.Recipients {
		o.Recipients[i].Status.SetOk()
	}

	for i := range o.Senders {
		o.Senders[i].SetOk()
	}
}
