package nodes

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

type DeviceReqOpts struct {
	Hostname string             `json:"host"`
	Device   string             `json:"device"`
	Class    string             `json:"class"`
	Status   status.BlockDevice `json:"status"`
}

type OsdReqOpts struct {
	Hostname string     `json:"host"`
	Id       string     `json:"id"`
	Device   string     `json:"device"`
	Reweight float64    `json:"rewight"`
	Status   status.Osd `json:"status"`
}

func (d *DeviceReqOpts) SetAdding() {
	d.Status.Desired = status.Added
	d.Status.IsProcessing = true
}

func (d *DeviceReqOpts) SetRemoving() {
	d.Status.Desired = status.Removed
	d.Status.IsProcessing = true
}

func (d *DeviceReqOpts) SetError() {
	d.Status.Current = status.Error
	d.Status.IsProcessing = false
}

func (d *DeviceReqOpts) SetCompleted() {
	d.Status.Current = status.Ok
	d.Status.IsProcessing = false
}

func (o *OsdReqOpts) SetError() {
	o.Status.Current = status.Error
}

func (o *OsdReqOpts) SetCompleted() {
	o.Status.Current = status.Ok
	o.Status.IsProcessing = false
}
