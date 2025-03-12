package tunings

import (
	"context"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addReqRecord(tuning definition.Tuning) {
	h := cubeMongo.GetGlobalHelper()
	err := h.UpdateOne(
		definition.TuningDB(),
		definition.TuningReqCollection(),
		bson.M{"id": tuning.Id},
		genUpsertPayload(tuning),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"failed to sync tuning record for %s (%s)",
			tuning.Name,
			err.Error(),
		)
	}
}

func genUpsertPayload(tuning definition.Tuning) bson.M {
	return bson.M{
		"$set": bson.M{
			"id":      tuning.Id,
			"name":    tuning.Name,
			"value":   tuning.Value,
			"enabled": tuning.Enabled,
			"status":  tuning.Status,
		},
	}
}

func updateTaskStatus(tuning *definition.Tuning) error {
	h := cubeMongo.GetGlobalHelper()
	return h.UpdateOne(
		definition.TuningDB(),
		definition.TuningReqCollection(),
		bson.M{"id": tuning.Id},
		bson.M{"$set": bson.M{"status.current": tuning.Status.Current}},
		options.Update().SetUpsert(true),
	)
}

func getUpdatingTunings() ([]definition.Tuning, error) {
	mongo := cubeMongo.GetGlobalHelper()
	cursor, err := mongo.GetQueryCursor(
		definition.TuningDB(),
		definition.TuningReqCollection(),
		bson.M{},
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	defer cursor.Close(ctx)
	return parseUpdatingTunings(cursor)
}

func parseUpdatingTunings(cursor *mongo.Cursor) ([]definition.Tuning, error) {
	tunings := []definition.Tuning{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	for cursor.Next(ctx) {
		tuning := definition.Tuning{}
		err := cursor.Decode(&tuning)
		if err != nil {
			return nil, err
		}

		log.Infof("tuning: %v", tuning)
		tunings = append(tunings, tuning)
	}

	return tunings, nil
}
