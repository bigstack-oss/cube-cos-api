package settings

import (
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/slack"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

const (
	Module        = "settings"
	DB            = "settings"
	ReqCollection = "requests"
	ReqTTL        = 3600
	PolicyDir     = "/etc/policies/alert_setting"
	PolicyV1      = "/etc/policies/alert_setting/alert_setting1_0.yml"

	MaxRecipientCount = 10
	MaxSlackCount     = 10
)

type Setting struct {
	Type             string `json:"type" bson:"type"`
	Key              string `json:"key" bson:"key"`
	Value            any    `json:"value" bson:"value"`
	IsReportRequired bool   `json:"-" bson:"-"`

	*TitlePrefix     `json:"titlePrefix,omitempty" bson:"titlePrefix,omitempty"`
	*email.Sender    `json:"sender,omitempty" bson:"sender,omitempty"`
	*email.Recipient `json:"recipient,omitempty" bson:"recipient,omitempty"`
	Slack            *slack.ApiChannel `json:"slack,omitempty" bson:"slack,omitempty"`

	Status status.Settings `json:"status" bson:"status"`
}

type TitlePrefix struct {
	Value  string          `json:"value" bson:"value"`
	Status status.Settings `json:"status" bson:"status"`
}

func (o *Setting) SetCreating() {
	o.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Created,
		IsUpdating: true,
	}
}

func (o *Setting) SetUpdating() {
	o.Status = status.Settings{
		Current:    status.Updating,
		Desired:    status.Updated,
		IsUpdating: true,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
	}
}

func (o *Setting) SetDeleting() {
	o.Status = status.Settings{
		Current:    status.Deleting,
		Desired:    status.Deleted,
		IsUpdating: true,
	}
}

func (o *Setting) SetError() {
	o.Status.Current = status.Error
	o.Status.IsUpdating = false
}

func (o *Setting) SetCompleted() {
	o.Status.Current = status.Ok
	o.Status.IsUpdating = false
}

func (o *Setting) GenTaskUpdate() Setting {
	return Setting{
		Type:   o.Type,
		Key:    o.Key,
		Status: o.Status,
	}
}
