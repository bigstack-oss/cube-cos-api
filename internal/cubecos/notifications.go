package cubecos

import (
	"context"

	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/notifications"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertNotification(notification notifications.Notification) error {
	h := bsmongo.GetGlobalHelper()
	err := h.Insert(
		notifications.Db,
		notifications.Toasts,
		notification,
	)
	if err != nil {
		log.Errorf("notifications: failed to insert notification(%v): %v", notification, err)
		return err
	}

	return nil
}

func ListNotifications(opts notifications.ListOpts) ([]notifications.Notification, error) {
	h := bsmongo.GetGlobalHelper()
	c, err := h.GetQueryCursor(
		notifications.Db,
		notifications.Toasts,
		genQueryFilter(opts),
		genQueryOpts(opts),
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

func genQueryFilter(opts notifications.ListOpts) bson.M {
	return bson.M{
		"time": bson.M{
			"$gte": opts.Start,
			"$lte": opts.Stop,
		},
	}
}

func genQueryOpts(opts notifications.ListOpts) *options.FindOptions {
	findOpts := options.Find().SetLimit(opts.Limit)
	if opts.Desending {
		return findOpts.SetSort(bson.D{{Key: "time", Value: -1}})
	}

	return findOpts
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
