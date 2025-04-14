package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord(titlePrefix setting.TitlePrefix) {
	err := mongo.GetGlobalHelper().UpdateOne(
		setting.DB,
		setting.ReqCollection,
		bson.M{"value": titlePrefix.Value},
		h.genUpsertPayload(titlePrefix),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"failed to sync title prefix record for %s (%s)",
			titlePrefix.Value,
			err.Error(),
		)
	}
}

func (h *helper) genUpsertPayload(titlePrefix setting.TitlePrefix) bson.M {
	return bson.M{
		"$set": bson.M{
			"type":  "titlePrefix",
			"value": titlePrefix.Value,
		},
	}
}
