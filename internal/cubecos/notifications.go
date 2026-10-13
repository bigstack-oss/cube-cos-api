package cubecos

import (
	"context"
	"fmt"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func InsertNotification(notification notifications.Notification) error {
	if notification.Id == "" {
		err := fmt.Errorf("notification id is required for insertion")
		log.Errorf("notifications: %v", err)
		return err
	}

	h := bsmongo.GetGlobalHelper()
	err := h.Insert(
		notifications.Db,
		notifications.ToastCollection,
		notification,
	)
	if err != nil {
		log.Errorf("notifications: failed to insert notification(%v): %v", notification, err)
		return err
	}

	return nil
}

func ListNotifications(opts notifications.ListOpts) ([]notifications.Notification, error) {
	c, err := bsmongo.GetGlobalHelper().Aggregate(
		notifications.Db,
		notifications.ToastCollection,
		genListPipeline(opts),
	)
	if err != nil {
		log.Errorf("notifications: failed to get query cursor for notifications: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(5))
	defer cancel()
	defer c.Close(ctx)
	return parseNotifications(c)
}

// note:
// reason to not converge this function with ListNotifications is that
// GetLastNotification is used to get the last notification only, so it does not need to consider the complicated query options like time fallback, sorting order,
// reading size and so on which makes this function simpler and more efficient.
func GetLastNotification() (*notifications.Notification, error) {
	c, err := bsmongo.GetGlobalHelper().Aggregate(
		notifications.Db,
		notifications.ToastCollection,
		genLastPipeline(),
	)
	if err != nil {
		log.Errorf("notifications: failed to get last notification cursor: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(60))
	defer cancel()
	defer c.Close(ctx)
	list, err := parseNotifications(c)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, nil
	}

	return &list[0], nil
}

// timeAtField holds the instant parsed from the string `time` while a pipeline
// runs; it is projected away before the documents are returned.
const timeAtField = "_timeAt"

// `time` is stored as a string whose offset depends on the node's zone when it
// was written (`Z`, `+00:00`, `+09:00`), so comparing or sorting the strings
// does not follow the instants. Parse it with $dateFromString first. A `time`
// that does not parse becomes null, which no date bound matches and which
// sorts last.
func genTimeAtStage() bson.D {
	return bson.D{{Key: "$addFields", Value: bson.D{{
		Key: timeAtField,
		Value: bson.D{{Key: "$dateFromString", Value: bson.D{
			{Key: "dateString", Value: "$time"},
			{Key: "onError", Value: nil},
			{Key: "onNull", Value: nil},
		}}},
	}}}}
}

func genDropTimeAtStage() bson.D {
	return bson.D{{Key: "$project", Value: bson.D{{Key: timeAtField, Value: 0}}}}
}

func genListPipeline(opts notifications.ListOpts) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		genTimeAtStage(),
		{{Key: "$match", Value: bson.D{{Key: timeAtField, Value: bson.D{
			{Key: "$gte", Value: opts.Start},
			{Key: "$lte", Value: opts.Stop},
		}}}}},
	}

	if opts.Desending {
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: timeAtField, Value: -1}}}})
	}

	if opts.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: opts.Limit}})
	}

	return append(pipeline, genDropTimeAtStage())
}

func genLastPipeline() mongo.Pipeline {
	return mongo.Pipeline{
		genTimeAtStage(),
		{{Key: "$sort", Value: bson.D{{Key: timeAtField, Value: -1}}}},
		{{Key: "$limit", Value: 1}},
		genDropTimeAtStage(),
	}
}

func parseNotifications(c *mongo.Cursor) ([]notifications.Notification, error) {
	list := []notifications.Notification{}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	for c.Next(ctx) {
		notification := notifications.Notification{}
		err := c.Decode(&notification)
		if err != nil {
			log.Errorf("notifications: failed to decode notification(%v)", err)
			return nil, err
		}

		list = append(
			list,
			notification,
		)
	}

	err := c.Err()
	if err != nil {
		log.Errorf("notifications: cursor has an error(%v)", err)
		return nil, err
	}

	return list, nil
}
