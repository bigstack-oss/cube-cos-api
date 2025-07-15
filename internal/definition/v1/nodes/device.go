package nodes

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"

type DeviceReqOpts struct {
	Hostname string             `json:"host"`
	Device   string             `json:"device"`
	Status   status.BlockDevice `json:"status"`
}

type OsdReqOpts struct {
	Hostname string     `json:"host"`
	Id       string     `json:"id"`
	Reweight float64    `json:"rewight"`
	Status   status.Osd `json:"status"`
}

func (d *DeviceReqOpts) SetError() {
	d.Status.Current = status.Error
	d.Status.IsAdding = false
	d.Status.IsRemoving = false
}

func (d *DeviceReqOpts) SetCompleted() {
	d.Status.Current = status.Ok
	d.Status.IsAdding = false
	d.Status.IsRemoving = false
}

func (o *OsdReqOpts) SetError() {
	o.Status.Current = status.Error
	o.Status.IsRestarting = false
	o.Status.IsRemoving = false
}

func (o *OsdReqOpts) SetCompleted() {
	o.Status.Current = status.Ok
	o.Status.IsRestarting = false
	o.Status.IsRemoving = false
}
