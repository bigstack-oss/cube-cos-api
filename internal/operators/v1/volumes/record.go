package volumes

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/volumes"
	"go.mongodb.org/mongo-driver/bson"
)

func (o *Operator) removePendingReqs() {
	h := mongo.GetGlobalHelper()
	h.DeleteAll(
		volumes.Db,
		volumes.ImageToVolumeReqCollection,
		bson.M{},
	)
}
