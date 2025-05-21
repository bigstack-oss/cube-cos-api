package cubecos

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func ForceRemovePendingReq(db, collection string) {
	h := mongo.GetGlobalHelper()
	err := h.DeleteAll(db, collection, bson.M{})
	if err != nil {
		log.Errorf("%s: failed to remove all pending requests: %v", db, err)
	}
}

func RemovePendingReq(db, collection string) {
	h := mongo.GetGlobalHelper()
	err := h.DeleteAll(
		db,
		collection,
		bson.M{"status.isUpdating": true},
	)
	if err != nil {
		log.Errorf("%s: failed to reset pending requests: %v", db, err)
	}
}
