package healths

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/health"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) setMoudleRepairingRecord() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.UpdateOne(
		health.Module,
		health.RepairingCollection,
		bson.M{"isRepairing": true},
		bson.M{"$set": bson.M{"type": "forceRepair", "module": h.moduleType, "isRepairing": true}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) setRepairingRecord() error {
	mongo := mongo.GetGlobalHelper()
	return mongo.UpdateOne(
		health.Module,
		health.RepairingCollection,
		bson.M{"isRepairing": true},
		bson.M{"$set": bson.M{"type": "checkRepair", "isRepairing": true}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) syncRepairingStatus(summary *cubecos.Health) {
	if h.hasCheckingAndRepairingRecord() {
		summary.Status.IsFixing = true
		summary.Status.Description = "Checking and repairing in progress"
	}
}

func (h *helper) hasCheckingAndRepairingRecord() bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		health.Module,
		health.RepairingCollection,
		bson.M{"isRepairing": true},
	)
	if err != nil {
		log.Errorf("healths(%s): failed to get count of repairing records: %v", queries.GetReqId(h.c), err)
		return false
	}

	return count > 0
}

func (h *helper) getModuleStatus() status.Health {
	s := status.Health{Current: status.Ok, IsFixing: false}
	if h.isModuleRepairing() {
		s.Current = status.Repairing
		s.IsFixing = true
		s.Description = "Repairing in progress"
	}

	return s
}

func (h *helper) isModuleRepairing() bool {
	mongo := mongo.GetGlobalHelper()
	count, err := mongo.GetCount(
		health.Module,
		health.RepairingCollection,
		bson.M{"module": h.moduleType, "isRepairing": true},
	)
	if err != nil {
		log.Errorf("healths(%s): failed to get count of repairing records: %v", queries.GetReqId(h.c), err)
		return false
	}

	return count > 0
}
