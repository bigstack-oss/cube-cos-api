package integrations

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addStorageModelReqRecord() {
	err := h.mongo.UpdateOne(
		storages.Db,
		storages.ModelReqCollection,
		bson.M{
			"model.driver": h.modelReqOpts.Model.Driver,
			"hostname":     h.modelReqOpts.Hostname,
			"reqId":        h.modelReqOpts.ReqId,
		},
		bson.M{"$set": h.modelReqOpts},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"integrations(%s): failed to add storage model request record(%v)",
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

	h.removeVerifiedRecordDuringUpdate()
}

func (h *helper) removeVerifiedRecordDuringUpdate() {
	if h.storageReqOpts.Status.Desired == status.Updated {
		h.removeVerificationRecord()
	}
}

func (h *helper) updateStorageTask() error {
	if h.storageReqOpts.Notify.IsNeeded {
		defer cubecos.InsertNotification(h.storageReqOpts.Notify.Payload)
	}

	if h.isVerificationReq() {
		h.updateStorageAsVerified(h.storageReqOpts.CinderDetails.Name)
	}

	return h.mongo.DeleteOne(
		storages.Db,
		storages.ReqCollection,
		bson.M{"reqId": h.storageReqOpts.ReqId},
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
			"hostname": h.modelReqOpts.Hostname,
			"reqId":    h.modelReqOpts.ReqId,
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
		bson.M{"cinderDetails.name": name},
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

func (h *helper) updateStorageAsVerified(name string) error {
	err := h.mongo.UpdateOne(
		storages.Db,
		storages.VerficationCollection,
		bson.M{"name": name},
		bson.M{"$set": bson.M{"name": name, "isVerified": true}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf("integrations(%s): failed to update storage(%s) as verified (%v)", h.reqId, name, err)
		return err
	}

	return nil
}

func (h *helper) syncVerifiedStorages(list *[]integration.Storage) {
	for i, storage := range *list {
		if storage.Type == "built-in" {
			(*list)[i].IsVerified = true
		}

		if h.hasVerifiedRecord(storage.Name) {
			(*list)[i].IsVerified = true
		}
	}
}

func (h *helper) hasVerifiedRecord(name string) bool {
	if name == storages.CubeStorage {
		return true
	}

	count, err := h.mongo.GetCount(
		storages.Db,
		storages.VerficationCollection,
		bson.M{"name": name, "isVerified": true},
	)
	if err != nil {
		log.Errorf("integrations(%s): failed to get storage verified status (%v)", h.reqId, err)
		return false
	}

	return count > 0
}

func (h *helper) removeVerificationRecord() {
	err := h.mongo.DeleteOne(
		storages.Db,
		storages.VerficationCollection,
		bson.M{"name": h.storageReqOpts.CinderDetails.Name},
	)
	if err != nil {
		log.Errorf(
			"integrations(%s): failed to remove storage(%s) verification record(%v)",
			h.reqId, h.storageReqOpts.CinderDetails.Name, err,
		)
	}
}
