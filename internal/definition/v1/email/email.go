package email

import (
	"net/mail"
)

const (
	emailSenders    = "emailSenders"
	emailRecipients = "emailRecipients"
)

type Options struct {
	Recipients []Recipient `json:"recipients" bson:"recipients"`
	Senders    []Sender    `json:"senders" bson:"senders"`
}

type Sender struct {
	Host     string `json:"host" bson:"host"`
	Port     int    `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Email    string `json:"from" bson:"from"`
}

type Recipient struct {
	Email string `json:"email" bson:"email"`
	Note  string `json:"note,omitempty" bson:"note,omitempty"`
}

func (r *Recipient) CheckFormat() error {
	_, err := mail.ParseAddress(r.Email)
	return err
}

func SenderCollection() string {
	return emailSenders
}

func RecipientCollection() string {
	return emailRecipients
}
