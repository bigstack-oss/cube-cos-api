package nodes

import (
	"strconv"
	"strings"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/blockdevice"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
)

type DeviceReqOpts struct {
	ReqId    string `json:"reqId" bson:"reqId"`
	Notify   `json:"notify" bson:"-"`
	Hostname string             `json:"hostname" bson:"hostname"`
	Device   string             `json:"device" bson:"device"`
	Class    string             `json:"class" bson:"class"`
	Status   status.BlockDevice `json:"status" `
}

type OsdReqOpts struct {
	Hostname string `json:"hostname" bson:"hostname"`
	OsdId    string `json:"osdId" bson:"osdId"`
	ReqId    string `json:"reqId" bson:"reqId"`
	Notify   `json:"notify" bson:"-"`
	Device   string     `json:"device" bson:"device"`
	Reweight float64    `json:"reweight" bson:"reweight"`
	Status   status.Osd `json:"status" bson:"status"`
}

func (d *DeviceReqOpts) SetAdding() {
	d.Status.Desired = status.Added
	d.Status.IsProcessing = true
}

func (d *DeviceReqOpts) SetUpdating() {
	d.Status.IsProcessing = true

	if strings.EqualFold(d.Class, blockdevice.SSD) {
		d.Status.Desired = status.Promoted
	}

	if strings.EqualFold(d.Class, blockdevice.HDD) {
		d.Status.Desired = status.Demoted
	}
}

func (d *DeviceReqOpts) SetRemoving() {
	d.Status.Desired = status.Removed
	d.Status.IsProcessing = true
}

func (d *DeviceReqOpts) SetError(msg string) {
	d.Status.Current = status.Error
	d.Status.IsProcessing = false

	d.SetDeviceNotification()
	d.Notify.Payload.AdditionalInfo["description"] = msg

	switch d.Status.Desired {
	case status.Added:
		d.Notify.Payload.Id = "DEV00001E"
	case status.Promoted:
		d.Notify.Payload.Id = "DEV00002E"
		d.Notify.Payload.AdditionalInfo["class"] = strings.ToUpper(d.Class)
	case status.Demoted:
		d.Notify.Payload.Id = "DEV00003E"
		d.Notify.Payload.AdditionalInfo["class"] = strings.ToUpper(d.Class)
	case status.Removed:
		d.Notify.Payload.Id = "DEV00004E"
	}
}

func (d *DeviceReqOpts) SetCompleted() {
	d.Status.Current = status.Ok
	d.Status.IsProcessing = false

	d.SetDeviceNotification()
	switch d.Status.Desired {
	case status.Added:
		d.Notify.Payload.Id = "DEV00001I"
	case status.Promoted:
		d.Notify.Payload.Id = "DEV00002I"
		d.Notify.Payload.AdditionalInfo["class"] = strings.ToUpper(d.Class)
	case status.Demoted:
		d.Notify.Payload.Id = "DEV00003I"
		d.Notify.Payload.AdditionalInfo["class"] = strings.ToUpper(d.Class)
	case status.Removed:
		d.Notify.Payload.Id = "DEV00003I"
	}
}

func (d *DeviceReqOpts) SetDeviceNotification() {
	d.Notify.Changes = true
	d.Notify.Payload = notifications.Notification{}
	d.Notify.Payload.NodeName = d.Hostname
	d.Notify.Payload.Time = time.LocalRFC3339(ostime.Now())
	d.Notify.Payload.AdditionalInfo = map[string]string{"device": d.Device}
}

func (o *OsdReqOpts) SetRestarting() {
	o.Status.Desired = status.Restarted
	o.Status.IsProcessing = true
}

func (o *OsdReqOpts) SetRemoving() {
	o.Status.Desired = status.Removed
	o.Status.IsProcessing = true
}

func (o *OsdReqOpts) SetReweighting() {
	o.Status.Desired = status.Reweighted
	o.Status.IsProcessing = true
}

func (o *OsdReqOpts) SetError() {
	o.Status.Current = status.Error
	o.Status.IsProcessing = false

	o.SetOsdNotification()
	switch o.Status.Desired {
	case status.Restarted:
		o.Notify.Payload.Id = "OSD00001E"
	case status.Reweighted:
		o.Notify.Payload.Id = "OSD00002E"
		o.Notify.Payload.AdditionalInfo["reweight"] = strconv.FormatFloat(o.Reweight, 'f', -1, 64)
	case status.Removed:
		o.Notify.Payload.Id = "OSD00003E"
	}
}

func (o *OsdReqOpts) SetCompleted() {
	o.Status.Current = status.Ok
	o.Status.IsProcessing = false

	o.SetOsdNotification()
	switch o.Status.Desired {
	case status.Restarted:
		o.Notify.Payload.Id = "OSD00001I"
	case status.Reweighted:
		o.Notify.Payload.Id = "OSD00002I"
		o.Notify.Payload.AdditionalInfo["reweight"] = strconv.FormatFloat(o.Reweight, 'f', -1, 64)
	case status.Removed:
		o.Notify.Payload.Id = "OSD00003I"
	}
}

func (d *OsdReqOpts) SetOsdNotification() {
	d.Notify.Changes = true
	d.Notify.Payload = notifications.Notification{}
	d.Notify.Payload.NodeName = d.Hostname
	d.Notify.Payload.Time = time.LocalRFC3339(ostime.Now())
	d.Notify.Payload.AdditionalInfo = map[string]string{"osdId": d.OsdId}
}
