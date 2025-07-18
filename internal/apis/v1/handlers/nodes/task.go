package nodes

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"go.mongodb.org/mongo-driver/bson"
)

func (h *helper) updateDeviceTask() error {
	changes.Add(nodes.Change{UseCacheInStream: false, Handler: h.handler})
	return h.mongo.DeleteOne(
		nodes.Db,
		nodes.ReqDeviceCollection,
		bson.M{
			"hostname": h.node,
			"device":   h.deviceReqOpts.Device,
		},
	)
}

func (h *helper) updateOsdTask() error {
	changes.Add(nodes.Change{UseCacheInStream: false, Handler: h.handler})
	return h.mongo.DeleteOne(
		nodes.Db,
		nodes.ReqOsdCollection,
		bson.M{
			"hostname": h.node,
			"device":   h.osdReqOpts.Device,
			"id":       h.osdReqOpts.Id,
		},
	)
}
