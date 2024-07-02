package tuning

import (
	"context"
	"time"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getTuningRecords() ([]definition.Tuning, error) {
	db := cubeMongo.GetGlobalHelper()
	defer db.Disconnect(context.Background())

	colls, err := db.GetAllCollections(definition.TuningDB())
	if err != nil {
		return nil, err
	}

	tunings := []definition.Tuning{}
	for _, coll := range colls {
		cursor, err := db.GetQueryCursor(definition.TuningDB(), coll, bson.M{})
		if err != nil {
			log.Errorf("failed to get cursor for %s (%s)", coll, err.Error())
			continue
		}

		appendTuningRecords(cursor, &tunings)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cursor.Close(ctx)
		cancel()
	}

	return tunings, nil
}

func appendTuningRecords(cursor *mongo.Cursor, tunings *[]definition.Tuning) {
	for cursor.Next(context.Background()) {
		tuning := definition.Tuning{}
		if err := cursor.Decode(&tuning); err != nil {
			log.Errorf("failed to decode tuning record (%s)", err.Error())
			continue
		}

		*tunings = append(*tunings, tuning)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate tuning cursor (%s)", cursor.Err().Error())
	}
}

func syncTuningRecord(tuning definition.Tuning) {
	db := cubeMongo.GetGlobalHelper()
	defer db.Disconnect(context.Background())

	filter := bson.M{"node.id": tuning.Node.ID, "name": tuning.Name}
	update := bson.M{"$set": tuning}
	err := db.UpdateOne(
		definition.TuningDB(),
		definition.TuningCollection(tuning.Name),
		filter,
		update,
		cubeMongo.CreateRecordIfNotExist,
	)
	if err != nil {
		log.Errorf(
			"failed to sync tuning record for %s (%s)",
			tuning.Name,
			err.Error(),
		)
	}
}

func updateRecordStatus(tuning *definition.Tuning) error {
	db := cubeMongo.GetGlobalHelper()
	defer db.Disconnect(context.Background())

	return db.UpdateOne(
		definition.TuningDB(),
		definition.TuningCollection(tuning.Name),
		bson.M{"node.id": tuning.Node.ID, "name": tuning.Name},
		tuning,
		options.Update().SetUpsert(true),
	)
}
