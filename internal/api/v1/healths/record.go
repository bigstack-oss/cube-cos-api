package healths

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func applyRepairRecord(health cubecos.Health) error {
	h := mongo.GetGlobalHelper()
	return h.UpdateOne(
		definition.HealthDB(),
		definition.RepairCollection(),
		bson.M{"dataCenter.name": health.DataCenter.Name},
		bson.M{"$set": health},
		options.Update().SetUpsert(true),
	)
}
