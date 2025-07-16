package nodes

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

func (h *helper) setDeviceCreateReq() nodes.DeviceReqOpts {
	return nodes.DeviceReqOpts{
		Device:   fmt.Sprintf("/dev/%s", h.device),
		Hostname: h.node,
		Status: status.BlockDevice{
			Desired:      status.Added,
			Current:      status.Adding,
			IsProcessing: true,
		},
	}
}

func (h *helper) setRemoveReqOpts() {
	h.deviceReqOpts.Device = fmt.Sprintf("/dev/%s", h.device)
	h.deviceReqOpts.Hostname = h.node
	h.deviceReqOpts.Status = status.BlockDevice{
		Desired:      status.Removed,
		Current:      status.Removing,
		IsProcessing: true,
	}
}
