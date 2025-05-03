package triggers

import (
	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord() {
	mongo := cubeMongo.GetGlobalHelper()
	err := mongo.UpdateOne(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": h.trigger.Name},
		h.genUpsertPayload(),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"failed to sync trigger record for %s (%s)",
			h.trigger.Name,
			err.Error(),
		)
	}
}

func (h *helper) updateTaskStatus() error {
	mongo := cubeMongo.GetGlobalHelper()
	return mongo.UpdateOne(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": h.trigger.Name},
		bson.M{"$set": bson.M{"status": h.trigger.Status}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) genUpsertPayload() bson.M {
	return bson.M{
		"$set": bson.M{
			"name":     h.trigger.Name,
			"match":    h.trigger.Match,
			"response": h.trigger.Response,
			"enabled":  h.trigger.Enabled,
			"status":   h.trigger.Status,
		},
	}
}

func (h *helper) hasUpdateHistory(t trigger.ApiOptions) bool {
	mongo := cubeMongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": t.Name},
	)
	if err != nil {
		return false
	}

	return count > 0
}

func (h *helper) getUpdateRecord(t trigger.ApiOptions) (*trigger.ApiOptions, error) {
	mongo := cubeMongo.GetGlobalHelper()
	pending, err := mongo.Get(
		trigger.DB,
		trigger.ReqCollection,
		bson.M{"name": t.Name},
	)
	if err != nil {
		return nil, err
	}

	record := &trigger.ApiOptions{}
	err = pending.Decode(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}
