package email

import (
	"fmt"
	"net/mail"
	"strconv"

	emailv2 "github.com/bigstack-oss/bigstack-dependency-go/pkg/email/v2"
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
	Auth     bool   `json:"auth" bson:"auth" yaml:"auth"`
	TLS      string `json:"tls" bson:"tls" yaml:"tls"`
}

func (c *CosSender) ToApiSchema() Sender {
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		port = 0
	}

	// a sender stored before tls existed was sent opportunistically, so keep
	// doing that rather than treating it as a new, unset request
	tls := emailv2.TLS(c.TLS)
	if tls == "" {
		tls = emailv2.TLSOpportunistic
	}

	auth := c.Auth
	return Sender{
		Host:     c.Host,
		Port:     port,
		Auth:     &auth,
		Username: c.Username,
		Password: &c.Password,
		TLS:      tls,
		From:     c.From,
	}
}

// HexSender is what hex_sdk alert_set_setting_sender_email takes. Every field
// is always present, so a missing value never reaches the policy as "null".
type HexSender struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Auth     bool   `json:"auth"`
	Username string `json:"username"`
	Password string `json:"password"`
	TLS      string `json:"tls"`
	From     string `json:"from"`
}

type Sender struct {
	Host           string           `json:"host,omitempty" bson:"host" yaml:"host,omitempty"`
	Port           int              `json:"port,omitempty" bson:"port" yaml:"port,omitempty"`
	Auth           *bool            `json:"auth,omitempty" bson:"auth" yaml:"auth,omitempty"`
	Username       string           `json:"username,omitempty" bson:"username" yaml:"username,omitempty"`
	Password       *string          `json:"password,omitzero" bson:"password" yaml:"password,omitempty"`
	TLS            emailv2.TLS      `json:"tls,omitempty" bson:"tls" yaml:"tls,omitempty"`
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

func (s *Sender) IsAuth() bool {
	return s.Auth != nil && *s.Auth
}

func (s *Sender) GetPassword() string {
	if s.Password == nil {
		return ""
	}

	return *s.Password
}

func IsValidTLS(tls emailv2.TLS) bool {
	switch tls {
	case emailv2.TLSNone, emailv2.TLSOpportunistic, emailv2.TLSMandatory:
		return true
	default:
		return false
	}
}

// Normalize fills auth and tls left unset by the request from current, the
// sender already in effect (nil when there is none). Without a current
// value, auth follows whether a username was given and tls is mandatory,
// matching pkg/email/v2, so an old client never silently downgrades.
// With auth off, credentials are dropped, since they are never sent.
func (s *Sender) Normalize(current *Sender) error {
	if s.Auth == nil {
		auth := s.Username != ""
		if current != nil && current.Auth != nil {
			auth = *current.Auth
		}
		s.Auth = &auth
	}

	if s.TLS == "" {
		s.TLS = emailv2.TLSMandatory
		if current != nil && current.TLS != "" {
			s.TLS = current.TLS
		}
	}

	if !IsValidTLS(s.TLS) {
		return fmt.Errorf(
			"email sender tls %q is invalid, want %q, %q or %q",
			s.TLS, emailv2.TLSNone, emailv2.TLSOpportunistic, emailv2.TLSMandatory,
		)
	}

	if !*s.Auth {
		empty := ""
		s.Username = ""
		s.Password = &empty
	}

	return nil
}

func (s *Sender) ToHexSchema() HexSender {
	return HexSender{
		Host:     s.Host,
		Port:     s.Port,
		Auth:     s.IsAuth(),
		Username: s.Username,
		Password: s.GetPassword(),
		TLS:      string(s.TLS),
		From:     s.From,
	}
}

func (s *Sender) EmailOptions() []emailv2.Option {
	opts := []emailv2.Option{
		emailv2.SenderHost(s.Host),
		emailv2.SenderPort(s.Port),
		emailv2.SenderFrom(s.From),
		emailv2.SenderTLS(s.TLS),
		emailv2.SenderAuth(s.IsAuth()),
	}
	if s.IsAuth() {
		opts = append(
			opts,
			emailv2.SenderUsername(s.Username),
			emailv2.SenderPassword(s.GetPassword()),
		)
	}

	return opts
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
