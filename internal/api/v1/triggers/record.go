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
	err := mongo.Insert(
		trigger.DB,
		trigger.Collection,
		h.trigger,
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
		trigger.Collection,
		bson.M{"id": h.trigger.Id},
		bson.M{"$set": bson.M{"status.current": h.trigger.Status.Current}},
		options.Update().SetUpsert(true),
	)
}
