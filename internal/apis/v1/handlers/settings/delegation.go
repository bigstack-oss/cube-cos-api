package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
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
			queries.GetReqId(h.c),
			req.Type,
			req.Key,
			err.Error(),
		)
	}
}

func (h *helper) genUpsertPayload(setting setting.Options) bson.M {
	switch setting.Type {
	case "titlePrefix":
		return genTitlePrefixUpdate(setting)
	case "emailSender":
		return genEmailSenderUpdate(setting)
	case "emailRecipient":
		return genEmailRecipientUpdate(setting)
	case "slackChannel":
		return genSlackChannelUpdate(setting)
	}

	return bson.M{}
}

func genTitlePrefixUpdate(setting setting.Options) bson.M {
	return bson.M{
		"$set": bson.M{
			"type":   setting.Type,
			"key":    setting.TitlePrefix.Value,
			"value":  setting.Value,
			"status": setting.Status,
		},
	}
}

func genEmailSenderUpdate(setting setting.Options) bson.M {
	return bson.M{
		"$set": bson.M{
			"type":   setting.Type,
			"key":    setting.Sender.Host,
			"sender": setting.Sender,
			"status": setting.Status,
		},
	}
}

func genEmailRecipientUpdate(setting setting.Options) bson.M {
	return bson.M{
		"$set": bson.M{
			"type":      setting.Type,
			"key":       setting.Key,
			"recipient": setting.Recipient,
			"status":    setting.Status,
		},
	}
}

func genSlackChannelUpdate(setting setting.Options) bson.M {
	return bson.M{
		"$set": bson.M{
			"type":   setting.Type,
			"key":    setting.Key,
			"slack":  setting.Slack,
			"status": setting.Status,
		},
	}
}
