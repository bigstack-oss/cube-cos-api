package integrations

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord() {
	err := h.mongo.UpdateOne(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{
			"hostname": base.Hostname,
			"name":     h.storageReqOpts.Name,
		},
		bson.M{"$set": h.storageReqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"integrations(%s): failed to add storage request record(%v)",
			h.reqId, err,
		)
	}
}

func (h *helper) updateStorageTask() error {
	return h.mongo.DeleteOne(
		storages.Db,
		storages.ReqCollection,
		bson.M{
			"hostname": h.storageReqOpts.Hostname,
			"name":     h.storageReqOpts.Name,
			"reqId":    h.storageReqOpts.ReqId,
		},
	)
}

func (h *helper) updateModelTask() error {
	return h.mongo.DeleteOne(
		storages.Db,
		storages.ModelReqCollection,
		bson.M{
			"hostname": h.modelReqOpts.Hostname,
			"vendor":   h.modelReqOpts.Vendor,
			"product":  h.modelReqOpts.Product,
			"reqId":    h.modelReqOpts.ReqId,
		},
	)
}
