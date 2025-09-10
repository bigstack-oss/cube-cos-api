package storages

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

const (
	Db                 = "storages"
	ReqCollection      = "storageRequests"
	ModelReqCollection = "modelRequests"
	CinderConf         = "/etc/cinder/cinder.conf"
)

type ReqOpts struct {
	ReqId    string `json:"reqId" bson:"reqId"`
	Name     string `json:"name" bson:"name"`
	Hostname string `json:"hostname" bson:"hostname"`
	Cinder   `json:"cinder" bson:"cinder"`
	Status   status.Storage `json:"status" bson:"status"`
}

type ModelReqOpts struct {
	ReqId    string `json:"reqId" bson:"reqId"`
	Name     string `json:"name" bson:"name"`
	Hostname string `json:"hostname" bson:"hostname"`
	Model    `json:"model" bson:"model"`
	Status   status.Model `json:"status" bson:"status"`
}

type Cinder struct {
	Name       string `json:"name" yaml:"name" bson:"name"`
	IsExternal bool   `json:"isExternal" yaml:"isExternal" bson:"isExternal"`
	Device     `json:"device" yaml:"device" bson:"device"`
	Storage    `json:"storage" yaml:"storage" bson:"storage"`
}

type Device struct {
	Vendor  string `json:"vendor" yaml:"vendor" bson:"vendor"`
	Product string `json:"product" yaml:"product" bson:"product"`
}

type Storage struct {
	Service    `json:"service" yaml:"service" bson:"service"`
	VolumeType `json:"volumeType" yaml:"volumeType" bson:"volumeType"`
	Image      `json:"image" yaml:"image" bson:"image"`
	UpdateTime string `json:"updateTime" yaml:"updateTime" bson:"updateTime"`
}

type Service struct {
	DriverSection []Conf         `json:"driverSection" yaml:"driverSection" bson:"driverSection"`
	ExtraSettings []ExtraSetting `json:"extraSettings" yaml:"extraSettings" bson:"extraSettings"`
}

type Conf struct {
	Key   string `json:"key" yaml:"key" bson:"key"`
	Value string `json:"value" yaml:"value" bson:"value"`
}

type ExtraSetting struct {
	SectionHeader string `json:"sectionHeader" yaml:"sectionHeader" bson:"sectionHeader"`
	Settings      []Conf `json:"settings" yaml:"settings" bson:"settings"`
}

type VolumeType struct {
	Settings []Conf `json:"settings" yaml:"settings" bson:"settings"`
}

type Image struct {
	UseMultipath   bool `json:"useMultipath" yaml:"useMultipath" bson:"useMultipath"`
	ForceMultipath bool `json:"forceMultipath" yaml:"forceMultipath" bson:"forceMultipath"`
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

func (r *ReqOpts) SetDeleting() {
	r.Status.Current = status.Deleting
	r.Status.Desired = status.Deleted
	r.Status.IsProcessing = true
}

func (r *ReqOpts) SetCompleted() {
	r.Status.Current = status.Completed
	r.Status.IsProcessing = false
}

func (r *ReqOpts) SetFailed() {
	r.Status.Current = status.Failed
	r.Status.IsProcessing = false
}

func (m *ModelReqOpts) SetCompleted() {
	m.Status.Current = status.Completed
	m.Status.IsProcessing = false
}

func (m *ModelReqOpts) SetFailed() {
	m.Status.Current = status.Failed
	m.Status.IsProcessing = false
}
