package nodes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	log "go-micro.dev/v5/logger"
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
		Id:                opts.ReqId,
		IsTaskInprogress:  false,
		NeedsNotification: true,
	})
}

func (h *helper) syncOsdChanges() {
	log.Infof("send osd changes to watchers")
	opts := h.osdReqOpts
	notifications.SetCacheById(opts.ReqId, opts.Notify.Payload)
	changes.Add(nodes.Change{
		Id:                opts.ReqId,
		IsTaskInprogress:  false,
		NeedsNotification: true,
	})
}
