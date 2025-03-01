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

func getTuningRecords() ([]definition.Tuning, error) {
	h := cubeMongo.GetGlobalHelper()
	colls, err := h.GetAllCollections(definition.TuningDB())
	if err != nil {
		return nil, err
	}

	tunings := []definition.Tuning{}
	for _, coll := range colls {
		cursor, err := h.GetQueryCursor(definition.TuningDB(), coll, bson.M{})
		if err != nil {
			log.Errorf("failed to get cursor for %s (%s)", coll, err.Error())
			continue
		}

		appendTuningRecords(cursor, &tunings)
		ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
		cursor.Close(ctx)
		cancel()
	}

	return tunings, nil
}

func appendTuningRecords(cursor *mongo.Cursor, tunings *[]definition.Tuning) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	for cursor.Next(ctx) {
		tuning := definition.Tuning{}
		err := cursor.Decode(&tuning)
		if err != nil {
			log.Errorf("failed to decode tuning record (%s)", err.Error())
			continue
		}

		*tunings = append(*tunings, tuning)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate tuning cursor (%s)", cursor.Err().Error())
	}
}

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
