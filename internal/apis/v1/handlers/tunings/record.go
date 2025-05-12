package tunings

import (
	"context"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord(tuning tunings.Tuning) {
	mongo := bsmongo.GetGlobalHelper()
	err := mongo.UpdateOne(
		tunings.DB(),
		tunings.ReqCollection(),
		bson.M{"id": tuning.Id},
		genUpsertPayload(tuning),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"tunings(%s): failed to sync tuning record for %s(%v)",
			h.reqId,
			tuning.Name,
			err,
		)
	}
}

func genUpsertPayload(tuning tunings.Tuning) bson.M {
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

func updateTaskStatus(tuning *tunings.Tuning) error {
	h := bsmongo.GetGlobalHelper()
	return h.UpdateOne(
		tunings.DB(),
		tunings.ReqCollection(),
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

func getUpdatingTunings() ([]tunings.Tuning, error) {
	mongo := bsmongo.GetGlobalHelper()
	cursor, err := mongo.GetQueryCursor(
		tunings.DB(),
		tunings.ReqCollection(),
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

func parseUpdatingTunings(cursor *mongo.Cursor) ([]tunings.Tuning, error) {
	list := []tunings.Tuning{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	for cursor.Next(ctx) {
		tuning := tunings.Tuning{}
		err := cursor.Decode(&tuning)
		if err != nil {
			return nil, err
		}

		list = append(list, tuning)
	}

	return list, nil
}
