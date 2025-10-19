package storages

import (
	"encoding/json"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

type Model struct {
	Driver    string `json:"driver" yaml:"driver" bson:"driver"`
	Vendor    string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Type      string `json:"type" yaml:"type" bson:"type"`
	Multipath []Path `json:"multipath" yaml:"multipath" bson:"multipath"`
	Storage   `json:"storage" yaml:"storage" bson:"storage"`
	Status    status.Model `json:"status" yaml:"status" bson:"status"`
}

type Path struct {
	Section     string       `json:"section" yaml:"section" bson:"section"`
	Attributes  []Attribute  `json:"attributes" yaml:"attributes" bson:"attributes"`
	SubSections []Subsection `json:"subSections" yaml:"subSections" bson:"subSections"`
}

type Subsection struct {
	Section    string      `json:"section" yaml:"section" bson:"section"`
	Attributes []Attribute `json:"attributes" yaml:"attributes" bson:"attributes"`
}

type Blacklist struct {
	Devnode string        `json:"devnode" yaml:"devnode" bson:"devnode"`
	Devices []ModelDevice `json:"devices" yaml:"devices" bson:"devices"`
}

type ModelDevice struct {
	Vendor   string      `json:"vendor" yaml:"vendor" bson:"vendor"`
	Product  string      `json:"product" yaml:"product" bson:"product"`
	Settings []Attribute `json:"settings" yaml:"settings" bson:"settings"`
}

type MultipathWwid struct {
	WWID     string      `json:"wwid" yaml:"wwid" bson:"wwid"`
	Settings []Attribute `json:"settings" yaml:"settings" bson:"settings"`
}

func (m *Model) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(b)
}

func (m *ModelReqOpts) SetCreating() {
	m.Status.Current = status.Creating
	m.Status.Desired = status.Created
	m.Status.IsProcessing = true
}

func (m *ModelReqOpts) SetUpdating() {
	m.Status.Current = status.Updating
	m.Status.Desired = status.Updated
	m.Status.IsProcessing = true
}

func (m *ModelReqOpts) SetDeleting() {
	m.Status.Current = status.Deleting
	m.Status.Desired = status.Deleted
	m.Status.IsProcessing = true
}

func (m *ModelReqOpts) SetCompleted() {
	m.Status.Current = status.Completed
	m.Status.IsProcessing = false

	m.SetModelNotification(true)
	switch m.Status.Desired {
	case status.Created:
		m.Notify.Payload.Id = "MDL00001I"
	case status.Updated:
		m.Notify.Payload.Id = "MDL00002I"
	case status.Deleted:
		m.Notify.Payload.Id = "MDL00003I"
	}
}

func (m *ModelReqOpts) SetFailed(msg string) {
	m.Status.Current = status.Failed
	m.Status.IsProcessing = false

	m.SetModelNotification(true)
	m.Notify.Payload.AdditionalInfo["description"] = msg

	switch m.Status.Desired {
	case status.Created:
		m.Notify.Payload.Id = "MDL00001E"
	case status.Updated:
		m.Notify.Payload.Id = "MDL00002E"
	case status.Deleted:
		m.Notify.Payload.Id = "MDL00003E"
	}
}

func (m *ModelReqOpts) SetModelNotification(shouldNotify bool) {
	m.Notify.IsNeeded = shouldNotify
	m.Notify.Payload = notifications.Notification{}
	m.Notify.Payload.NodeName = m.Hostname
	m.Notify.Payload.Time = time.LocalRFC3339(ostime.Now())
	m.Notify.Payload.AdditionalInfo = map[string]string{
		"type":   m.Type,
		"vendor": m.Vendor,
		"driver": m.Driver,
	}
}
