package storages

import (
	"encoding/json"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

const (
	Db                 = "storages"
	ReqCollection      = "storageRequests"
	ModelReqCollection = "modelRequests"
	CinderConf         = "/etc/cinder/cinder.conf"

	TmpUploadedStorageModel     = "/tmp/storage-model.yaml"
	TmpUploadedStorageModelList = "/tmp/storage-model-list.yaml"
	TmpUploadedStorage          = "/tmp/storage.yaml"
	TmpUploadedStorageList      = "/tmp/storage-list.yaml"

	DefaultType = "__DEFAULT__"
)

type ReqOpts struct {
	ReqId         string `json:"reqId" bson:"reqId"`
	Name          string `json:"name" bson:"name"`
	Hostname      string `json:"hostname" bson:"hostname"`
	CinderDetails `json:"cinderDetails" bson:"cinderDetails"`
	Status        status.Storage `json:"status" bson:"status"`
	Notify        `json:"notify" bson:"notify"`
}

type ModelReqOpts struct {
	ReqId    string `json:"reqId" bson:"reqId"`
	Hostname string `json:"hostname" bson:"hostname"`
	Model    `json:"model" bson:"model"`
	Status   status.Model `json:"status" bson:"status"`
	Notify   `json:"notify" bson:"notify"`
}

type Cinder struct {
	Name       string `json:"name" yaml:"name" bson:"name"`
	Driver     string `json:"driver" yaml:"driver" bson:"driver"`
	Vendor     string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Model      string `json:"model" yaml:"model" bson:"model"`
	IsDefault  bool   `json:"isDefault" yaml:"isDefault" bson:"isDefault"`
	IsBuiltIn  bool   `json:"isBuiltIn" yaml:"isBuiltIn" bson:"isBuiltIn"`
	UpdateTime string `json:"updateTime" yaml:"updateTime" bson:"updateTime"`
}

type CinderDetails struct {
	Name       string `json:"name" yaml:"name" bson:"name"`
	Driver     string `json:"driver" yaml:"driver" bson:"driver"`
	Vendor     string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Model      string `json:"model" yaml:"model" bson:"model"`
	IsDefault  bool   `json:"isDefault" yaml:"isDefault" bson:"isDefault"`
	IsBuiltIn  bool   `json:"isBuiltIn" yaml:"isBuiltIn" bson:"isBuiltIn"`
	UpdateTime string `json:"updateTime" yaml:"updateTime" bson:"updateTime"`
	Storage    `json:"storage" yaml:"storage" bson:"storage"`
}

type Device struct {
	Driver string `json:"driver" yaml:"driver" bson:"driver"`
}

type Storage struct {
	Service    `json:"service" yaml:"service" bson:"service"`
	VolumeType `json:"volumeType" yaml:"volumeType" bson:"volumeType"`
	Image      `json:"image" yaml:"image" bson:"image"`
	UpdateTime string `json:"updateTime" yaml:"updateTime" bson:"updateTime"`
}

type Service struct {
	DriverSection    []Attribute       `json:"driverSection" yaml:"driverSection" bson:"driverSection"`
	ExtraSettings    []ExtraSetting    `json:"extraSettings" yaml:"extraSettings" bson:"extraSettings"`
	ExtraConfigFiles []ExtraConfigFile `json:"extraConfigFiles" yaml:"extraConfigFiles" bson:"extraConfigFiles"`
}

type Attribute struct {
	Key   string `json:"key" yaml:"key" bson:"key"`
	Value string `json:"value" yaml:"value" bson:"value"`
}

type ExtraSetting struct {
	SectionHeader string      `json:"sectionHeader" yaml:"sectionHeader" bson:"sectionHeader"`
	Settings      []Attribute `json:"settings" yaml:"settings" bson:"settings"`
}

type ExtraConfigFile struct {
	Name    string `json:"name" yaml:"name" bson:"name"`
	Content string `json:"content" yaml:"content" bson:"content"`
}

type VolumeType struct {
	Settings []Attribute `json:"settings" yaml:"settings" bson:"settings"`
}

type Image struct {
	UseMultipath   bool `json:"useMultipath" yaml:"useMultipath" bson:"useMultipath"`
	ForceMultipath bool `json:"forceMultipath" yaml:"forceMultipath" bson:"forceMultipath"`
}

type VerficationResult struct {
	IsCinderServiceUp      bool `json:"isCinderServiceUp"`
	IsTestVolumeSuccessful bool `json:"isTestVolumeSuccessful"`
}

type Notify struct {
	IsNeeded bool                       `json:"isNeeded" bson:"isNeeded"`
	Payload  notifications.Notification `json:"payload" bson:"payload"`
}

func (r *ReqOpts) SetCreating() {
	r.Status.Current = status.Creating
	r.Status.Desired = status.Created
	r.Status.IsProcessing = true
}

func (r *ReqOpts) SetUpdating() {
	r.Status.Current = status.Updating
	r.Status.Desired = status.Updated
	r.Status.IsProcessing = true
}

func (r *ReqOpts) SetSettingAsDefault() {
	r.Status.Current = status.SettingToDefault
	r.Status.Desired = status.Defaulted
	r.Status.IsProcessing = true
}

func (r *ReqOpts) SetDeleting() {
	r.Status.Current = status.Deleting
	r.Status.Desired = status.Deleted
	r.Status.IsProcessing = true
}

func (r *ReqOpts) SetCompleted() {
	r.Status.Current = status.Completed
	r.Status.IsProcessing = false

	r.SetStorageNotification(true)
	switch r.Status.Desired {
	case status.Created:
		r.Notify.Payload.Id = "STG00001I"
	case status.Updated:
		r.Notify.Payload.Id = "STG00002I"
	case status.Deleted:
		r.Notify.Payload.Id = "STG00003I"
	case status.Defaulted:
		r.SetStorageNotification(false)
	}
}

func (r *ReqOpts) SetFailed(msg string) {
	r.Status.Current = status.Failed
	r.Status.IsProcessing = false

	r.SetStorageNotification(true)
	r.Notify.Payload.AdditionalInfo["description"] = msg

	switch r.Status.Desired {
	case status.Created:
		r.Notify.Payload.Id = "STG00001E"
	case status.Updated:
		r.Notify.Payload.Id = "STG00002E"
	case status.Deleted:
		r.Notify.Payload.Id = "STG00003E"
	case status.Defaulted:
		r.SetStorageNotification(false)
	}
}

func (r *ReqOpts) SetStorageNotification(shouldNotify bool) {
	r.Notify.IsNeeded = shouldNotify
	r.Notify.Payload = notifications.Notification{}
	r.Notify.Payload.NodeName = r.Hostname
	r.Notify.Payload.Time = time.LocalRFC3339(ostime.Now())
	r.Notify.Payload.AdditionalInfo = map[string]string{"name": r.Name}
}

func (m *ModelReqOpts) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(b)
}
