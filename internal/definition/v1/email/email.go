package email

import (
	"fmt"
	"net/mail"
	"net/smtp"
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
	Host           string `json:"host" bson:"host"`
	Port           int    `json:"port" bson:"port"`
	Username       string `json:"username" bson:"username"`
	Password       string `json:"password" bson:"password"`
	Email          string `json:"email" bson:"email"`
	AccessVerified bool   `json:"accessVerified" bson:"accessVerified"`
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

type Recipient struct {
	Email string `json:"email" bson:"email"`
	Note  string `json:"note" bson:"note"`
}

func (r *Recipient) CheckEmailFormat() error {
	_, err := mail.ParseAddress(r.Email)
	return err
}
