package supportfiles

import (
	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addReqRecord(file support.File) {
	h := cubeMongo.GetGlobalHelper()
	err := h.UpdateOne(
		support.FileDB,
		support.FileReqCollection,
		genFilter(file),
		genUpsertPayload(file),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"failed to sync tuning record for %s (%s)",
			file.Name,
			err.Error(),
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
	mongo := cubeMongo.GetGlobalHelper()
	return mongo.UpdateOne(
		support.FileDB,
		support.FileReqCollection,
		h.genTaskFilter(),
		bson.M{"$set": bson.M{"status.current": h.file.Status.Current}},
		options.Update().SetUpsert(true),
	)
}

func (h *helper) genTaskFilter() bson.M {
	return bson.M{
		"group":            h.file.Group,
		"source.host":      h.file.Source.Host,
		"status.createdAt": h.file.Status.CreatedAt,
	}
}
