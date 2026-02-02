package images

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"go.mongodb.org/mongo-driver/bson"
)

func (o *Operator) removePendingReqs() {
	h := mongo.GetGlobalHelper()
	h.DeleteAll(
		images.Db,
		images.ReqCollection,
		bson.M{},
	)
}

func (o *Operator) housekeepDanglingFinishedReqs() {
	for {
		wait.Seconds(60)
		o.removeDanglingImportedReqs()
	}
}

func (o *Operator) removeDanglingImportedReqs() {
	h := mongo.GetGlobalHelper()
	h.DeleteAll(
		images.Db,
		images.ReqCollection,
		bson.M{
			"status.current":        "importing",
			"status.processPercent": 100,
		},
	)
}
