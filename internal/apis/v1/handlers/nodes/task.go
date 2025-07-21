package nodes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) updateDeviceTask() error {
	defer h.syncDeviceChanges()
	return h.mongo.DeleteOne(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{
			"hostname": h.deviceReqOpts.Hostname,
			"device":   h.deviceReqOpts.Device,
			"reqId":    h.deviceReqOpts.ReqId,
		},
	)
}

func (h *helper) updateOsdTask() error {
	defer h.syncOsdChanges()
	return h.mongo.DeleteOne(
		nodes.Db,
		nodes.ReqOsdCollection,
		bson.M{
			"hostname": h.osdReqOpts.Hostname,
			"device":   h.osdReqOpts.Device,
			"reqId":    h.osdReqOpts.ReqId,
			"osdId":    h.osdReqOpts.OsdId,
		},
	)
}

func (h *helper) syncDeviceChanges() {
	opts := h.deviceReqOpts
	notifications.SetCacheById(opts.ReqId, opts.Notify.Payload)
	changes.Add(nodes.Change{
		Id:                h.deviceReqOpts.ReqId,
		IsTaskInprogress:  false,
		NeedsNotification: opts.Notify.Changes,
	})
}

func (h *helper) syncOsdChanges() {
	opts := h.osdReqOpts
	notifications.SetCacheById(opts.ReqId, opts.Notify.Payload)
	changes.Add(nodes.Change{
		Id:                h.osdReqOpts.ReqId,
		IsTaskInprogress:  false,
		NeedsNotification: opts.Notify.Changes,
	})
}
