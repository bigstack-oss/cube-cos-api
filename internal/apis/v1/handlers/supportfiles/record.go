package supportfiles

import (
	bsmongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord(file support.File) {
	mongo := bsmongo.GetGlobalHelper()
	err := mongo.UpdateOne(
		support.FileDB,
		support.FileReqCollection,
		genFilter(file),
		genUpsertPayload(file),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"supportfiles(%s): failed to sync tuning record for %s(%v)",
			h.reqId,
			file.Name,
			err,
		)
	}
}

func genFilter(file support.File) bson.M {
	return bson.M{
		"group":            file.Group,
		"source.host":      file.Source.Host,
		"status.createdAt": file.Status.CreatedAt,
	}
}

func genUpsertPayload(file support.File) bson.M {
	return bson.M{
		"$set": bson.M{
			"group":  file.Group,
			"source": file.Source,
			"status": file.Status,
		},
	}
}

func (h *helper) updateSupportFileTask() error {
	mongo := bsmongo.GetGlobalHelper()
	return mongo.DeleteOne(
		support.FileDB,
		support.FileReqCollection,
		genTaskFilter(h.file),
	)
}

func genTaskFilter(file support.File) bson.M {
	return bson.M{
		"group":            file.Group,
		"source.host":      file.Source.Host,
		"status.createdAt": file.Status.CreatedAt,
	}
}
