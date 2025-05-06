package healths

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) setRepairingRecord() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.UpdateOne(
		v1.Healths,
		v1.HealthRepairingCollection,
		bson.M{"isRepairing": true},
		bson.M{"$set": bson.M{"isRepairing": true}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) syncRepairingStatus(summary *cubecos.Health) {
	if h.hasCheckingAndRepairingRecord() {
		summary.Status.Current = status.Ok
		summary.Status.IsFixing = true
		summary.Status.Description = "Checking and repairing in progress"
	}
}

func (h *helper) hasCheckingAndRepairingRecord() bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		v1.Healths,
		v1.HealthRepairingCollection,
		bson.M{"isRepairing": true},
	)
	if err != nil {
		log.Errorf("healths(%s): failed to get count of repairing records: %v", api.GetReqId(h.c), err)
		return false
	}

	return count > 0
}
