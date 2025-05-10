package tunings

import (
	"context"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addReqRecord(tuning v1.Tuning) {
	h := cubeMongo.GetGlobalHelper()
	err := h.UpdateOne(
		v1.TuningDB(),
		v1.TuningReqCollection(),
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

func genUpsertPayload(tuning v1.Tuning) bson.M {
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

func updateTaskStatus(tuning *v1.Tuning) error {
	h := cubeMongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.TuningDB(),
		v1.TuningReqCollection(),
		bson.M{"id": tuning.Id},
		bson.M{
			"$set": bson.M{
				"status.current":    tuning.Status.Current,
				"status.isUpdating": false,
			},
		},
		options.Update().SetUpsert(true),
	)
}

func getUpdatingTunings() ([]v1.Tuning, error) {
	mongo := cubeMongo.GetGlobalHelper()
	cursor, err := mongo.GetQueryCursor(
		v1.TuningDB(),
		v1.TuningReqCollection(),
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

func parseUpdatingTunings(cursor *mongo.Cursor) ([]v1.Tuning, error) {
	tunings := []v1.Tuning{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	for cursor.Next(ctx) {
		tuning := v1.Tuning{}
		err := cursor.Decode(&tuning)
		if err != nil {
			return nil, err
		}

		tunings = append(tunings, tuning)
	}

	return tunings, nil
}
