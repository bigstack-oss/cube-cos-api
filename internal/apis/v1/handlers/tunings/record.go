package tunings

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) updateRecord(host string) {
	err := h.mongo.UpdateOne(
		tunings.DB(),
		tunings.ReqCollection(),
		bson.M{"name": h.tuning.Name, "host": host},
		h.genUpsertPayload(),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"tunings(%s): failed to sync record for %s(%v)",
			h.reqId,
			h.tuning.Name,
			err,
		)
	}
}

func (h *helper) genUpsertPayload() bson.M {
	return bson.M{
		"$set": bson.M{
			"name":    h.tuning.Name,
			"value":   h.tuning.Value,
			"enabled": h.tuning.Enabled,
			"status":  h.tuning.Status,
		},
	}
}

func (h *helper) updateTaskStatus() error {
	return h.mongo.UpdateOne(
		tunings.DB(),
		tunings.ReqCollection(),
		bson.M{"name": h.tuning.Name, "host": h.tuning.Node.Hostname},
		bson.M{
			"$set": bson.M{
				"status.current":    h.tuning.Status.Current,
				"status.isUpdating": false,
			},
		},
		options.Update().SetUpsert(true),
	)
}
