package tunings

import (
	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addReqRecord(tuning definition.Tuning) {
	h := cubeMongo.GetGlobalHelper()
	err := h.Insert(
		definition.TuningDB(),
		definition.TuningReqCollection(),
		tuning,
	)
	if err != nil {
		log.Errorf(
			"failed to sync tuning record for %s (%s)",
			tuning.Name,
			err.Error(),
		)
	}
}

func updateTaskStatus(tuning *definition.Tuning) error {
	mongo := cubeMongo.GetGlobalHelper()
	return mongo.UpdateOne(
		definition.TuningDB(),
		definition.TuningReqCollection(),
		bson.M{"id": tuning.Id},
		bson.M{"$set": bson.M{"status.current": tuning.Status.Current}},
		options.Update().SetUpsert(true),
	)
}
