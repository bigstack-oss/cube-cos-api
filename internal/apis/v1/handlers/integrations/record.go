package integrations

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addStorageModelReqRecord() {
	err := h.mongo.UpdateOne(
		storages.Db,
		storages.ReqCollection,
		bson.M{
			"hostname":           base.Hostname,
			"cinderDetails.name": h.storageReqOpts.CinderDetails.Name,
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

func (h *helper) addStorageReqRecord() {
	err := h.mongo.UpdateOne(
		storages.Db,
		storages.ReqCollection,
		bson.M{
			"hostname":           base.Hostname,
			"cinderDetails.name": h.storageReqOpts.CinderDetails.Name,
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
	if h.storageReqOpts.Notify.IsNeeded {
		defer cubecos.InsertNotification(h.storageReqOpts.Notify.Payload)
	}

	log.Infof("-------")
	log.Infof(h.storageReqOpts.Hostname)
	log.Infof(h.storageReqOpts.CinderDetails.Name)
	log.Infof(h.storageReqOpts.ReqId)

	return h.mongo.DeleteOne(
		storages.Db,
		storages.ReqCollection,
		bson.M{
			"hostname":           h.storageReqOpts.Hostname,
			"cinderDetails.Name": h.storageReqOpts.CinderDetails.Name,
			"reqId":              h.storageReqOpts.ReqId,
		},
	)
}

func (h *helper) updateModelTask() error {
	if h.modelReqOpts.Notify.IsNeeded {
		defer cubecos.InsertNotification(h.modelReqOpts.Notify.Payload)
	}

	return h.mongo.DeleteOne(
		storages.Db,
		storages.ModelReqCollection,
		bson.M{
			"hostname":     h.modelReqOpts.Hostname,
			"model.driver": h.modelReqOpts.Driver,
			"reqId":        h.modelReqOpts.ReqId,
		},
	)
}

func (h *helper) syncProcessingStorages(storages *[]integration.Storage) {
	for i, storage := range *storages {
		if h.isStorageProcessing(storage.Name) {
			h.syncStorageProcessingStatus(&(*storages)[i])
		}
	}
}

func (h *helper) isStorageProcessing(name string) bool {
	count, err := h.mongo.GetCount(
		storages.Db,
		storages.ReqCollection,
		bson.M{
			"cinderDetails.name": name,
			"hostname":           base.Hostname,
		},
	)
	if err != nil {
		log.Errorf("integrations(%s): failed to get storage processing status (%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) syncStorageProcessingStatus(storage *integration.Storage) {
	reqOpts := &storages.ReqOpts{}
	doc, err := h.mongo.Get(
		storages.Db,
		storages.ReqCollection,
		bson.M{"cinderDetails.name": storage.Name},
	)
	if err != nil {
		log.Errorf("integrations(%s): failed to get storage request record (%v)", h.reqId, err)
		return
	}
	if doc == nil {
		return
	}

	err = doc.Decode(reqOpts)
	if err != nil {
		log.Errorf("integrations(%s): failed to decode storage request record (%v)", h.reqId, err)
		return
	}

	storage.Status = reqOpts.Status
}

func (h *helper) syncProcessingModels(models *[]storages.Model) {
	for i, model := range *models {
		if h.isModelProcessing(model.Driver) {
			h.syncModelProcessingStatus(&(*models)[i])
		}
	}
}

func (h *helper) isModelProcessing(driver string) bool {
	count, err := h.mongo.GetCount(
		storages.Db,
		storages.ModelReqCollection,
		bson.M{"driver": driver},
	)
	if err != nil {
		log.Errorf("integrations(%s): failed to get storage processing status (%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) syncModelProcessingStatus(model *storages.Model) {
	reqOpts := &storages.ModelReqOpts{}
	doc, err := h.mongo.Get(
		storages.Db,
		storages.ModelReqCollection,
		bson.M{"model.driver": model.Driver},
	)
	if err != nil {
		log.Errorf("integrations(%s): failed to get storage request record (%v)", h.reqId, err)
		return
	}
	if doc == nil {
		return
	}

	err = doc.Decode(reqOpts)
	if err != nil {
		log.Errorf("integrations(%s): failed to decode storage request record (%v)", h.reqId, err)
		return
	}

	model.Status = reqOpts.Status
}
