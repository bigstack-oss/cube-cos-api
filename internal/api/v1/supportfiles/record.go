package supportfiles

import (
	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addReqRecord(supportFile v1.SupportFile) {
	h := cubeMongo.GetGlobalHelper()
	err := h.UpdateOne(
		v1.SupportFileDB,
		v1.SupportFileReqCollection,
		bson.M{"id": supportFile.Id},
		genUpsertPayload(supportFile),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"failed to sync tuning record for %s (%s)",
			supportFile.Name,
			err.Error(),
		)
	}
}

func genUpsertPayload(supportFile v1.SupportFile) bson.M {
	return bson.M{
		"$set": bson.M{
			"id":      supportFile.Id,
			"name":    supportFile.Name,
			"comment": supportFile.Comment,
			"status":  supportFile.Status,
		},
	}
}
