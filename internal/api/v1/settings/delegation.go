package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *helper) addReqRecord(req setting.Options) {
	err := mongo.GetGlobalHelper().UpdateOne(
		setting.DB,
		setting.ReqCollection,
		bson.M{"type": req.Type},
		h.genUpsertPayload(req),
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Errorf(
			"settings(%s): failed to sync %s record for %s (%s)",
			api.GetReqId(h.c),
			req.Type,
			req.GetKey(),
			err.Error(),
		)
	}
}

func (h *helper) genUpsertPayload(setting setting.Options) bson.M {
	switch setting.Type {
	case "titlePrefix":
		return bson.M{
			"$set": bson.M{
				"type":   setting.Type,
				"key":    setting.GetKey(),
				"value":  setting.Value,
				"status": setting.Status,
			},
		}
	case "emailSender":
		return bson.M{
			"$set": bson.M{
				"type":   setting.Type,
				"key":    setting.GetKey(),
				"sender": setting.Sender,
				"status": setting.Status,
			},
		}
	case "emailRecipient":
		return bson.M{
			"$set": bson.M{
				"type":      setting.Type,
				"key":       setting.GetKey(),
				"recipient": setting.Recipient,
				"status":    setting.Status,
			},
		}
	}

	return bson.M{
		"$set": bson.M{
			"type":   setting.Type,
			"key":    setting.GetKey(),
			"value":  setting.Value,
			"status": setting.Status,
		},
	}
}
