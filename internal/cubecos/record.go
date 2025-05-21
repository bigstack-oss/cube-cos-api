package cubecos

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
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

func RemoveHostPendingReq(db, collection string) {
	h := mongo.GetGlobalHelper()
	err := h.DeleteAll(
		db,
		collection,
		bson.M{
			"host":              base.Hostname,
			"status.isUpdating": true,
		},
	)
	if err != nil {
		log.Errorf("%s: failed to reset pending requests: %v", db, err)
	}
}
