package settings

import (
	"context"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/google/uuid"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getSettingRecord() (*v1.Setting, error) {
	titlePrefix, err := getTitlePrefix()
	if err != nil {
		log.Errorf("failed to get title prefix (%s)", err.Error())
		return nil, err
	}

	senders, err := getEmailSenderRecords()
	if err != nil {
		log.Errorf("failed to get email sender records (%s)", err.Error())
		return nil, err
	}

	recipients, err := getEmailRecipientRecords()
	if err != nil {
		log.Errorf("failed to get email recipient records (%s)", err.Error())
		return nil, err
	}

	webhooks, err := getSlackWebhookRecords()
	if err != nil {
		log.Errorf("failed to get slack webhook records (%s)", err.Error())
		return nil, err
	}

	return &v1.Setting{
		TitlePrefix: titlePrefix,
		Email: v1.Email{
			Senders:    senders,
			Recipients: recipients,
		},
		Slack: v1.Slack{
			Webhooks: webhooks,
		},
	}, nil
}

func upsertTitlePrefixRecord(titlePrefix string) error {
	h := cubeMongo.GetGlobalHelper()
	opts := options.Update().SetUpsert(true)
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.TitlePrefixCollection(),
		bson.M{},
		bson.M{"$set": bson.M{"value": titlePrefix}},
		opts,
	)
}

func upsertEmailSenderRecord(emailSender v1.EmailSender) error {
	h := cubeMongo.GetGlobalHelper()
	opts := options.Update().SetUpsert(true)
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailSenderCollection(),
		bson.M{},
		bson.M{"$set": emailSender},
		opts,
	)
}

func getTitlePrefix() (string, error) {
	h := cubeMongo.GetGlobalHelper()
	cursor, err := h.GetQueryCursor(v1.SettingsDB(), v1.TitlePrefixCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get cursor for email sender (%s)", err.Error())
		return "", err
	}
	curCtx, curCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer curCancel()
	defer cursor.Close(curCtx)

	nxtCtx, nxtCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer nxtCancel()
	for cursor.Next(nxtCtx) {
		titlePrefix := v1.TitlePrefix{}
		if err := cursor.Decode(&titlePrefix); err != nil {
			continue
		}
		return titlePrefix.Value, nil
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email sender records (%s)", cursor.Err().Error())
	}

	return "", nil
}

func getEmailSenderRecords() ([]v1.EmailSender, error) {
	h := cubeMongo.GetGlobalHelper()
	senders := []v1.EmailSender{}
	cursor, err := h.GetQueryCursor(v1.SettingsDB(), v1.EmailSenderCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get cursor for email sender (%s)", err.Error())
		return senders, err
	}
	curCtx, curCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer curCancel()
	defer cursor.Close(curCtx)

	nxtCtx, nxtCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer nxtCancel()
	for cursor.Next(nxtCtx) {
		sender := v1.EmailSender{}
		if err := cursor.Decode(&sender); err != nil {
			continue
		}
		senders = append(senders, sender)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email sender records (%s)", cursor.Err().Error())
	}

	return senders, nil
}

func deleteEmailSenderRecord() error {
	h := cubeMongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailSenderCollection(),
		bson.M{},
		bson.M{"$set": bson.M{"deleted": true}},
	)
}

func createEmailRecipientRecord(emailRecipient v1.EmailRecipient) error {
	h := cubeMongo.GetGlobalHelper()
	emailRecipient.ID = uuid.NewString()
	return h.Insert(
		v1.SettingsDB(),
		v1.EmailRecipientCollection(),
		emailRecipient,
	)
}

func getEmailRecipientRecords() ([]v1.EmailRecipient, error) {
	h := cubeMongo.GetGlobalHelper()
	recipients := []v1.EmailRecipient{}
	cursor, err := h.GetQueryCursor(v1.SettingsDB(), v1.EmailRecipientCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get cursor for email recipient (%s)", err.Error())
		return recipients, err
	}
	curCtx, curCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer curCancel()
	defer cursor.Close(curCtx)

	nxtCtx, nxtCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer nxtCancel()
	for cursor.Next(nxtCtx) {
		recipient := v1.EmailRecipient{}
		if err := cursor.Decode(&recipient); err != nil {
			continue
		}
		recipients = append(recipients, recipient)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate email recipient records (%s)", cursor.Err().Error())
	}

	return recipients, nil
}

func updateEmailRecipientRecord(emailRecipient v1.EmailRecipient) error {
	h := cubeMongo.GetGlobalHelper()
	filter := bson.M{"id": emailRecipient.ID}
	update := bson.M{"$set": emailRecipient}
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailRecipientCollection(),
		filter,
		update,
	)
}

func deleteEmailRecipientRecord(id string) error {
	h := cubeMongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailRecipientCollection(),
		bson.M{"id": id},
		bson.M{"$set": bson.M{"deleted": true}},
	)
}

func createSlackWebhookRecord(webhook v1.SlackWebhook) error {
	h := cubeMongo.GetGlobalHelper()
	webhook.ID = uuid.NewString()
	return h.Insert(
		v1.SettingsDB(),
		v1.SlackWebhookCollection(),
		webhook,
	)
}

func getSlackWebhookRecords() ([]v1.SlackWebhook, error) {
	h := cubeMongo.GetGlobalHelper()
	webhooks := []v1.SlackWebhook{}
	cursor, err := h.GetQueryCursor(v1.SettingsDB(), v1.SlackWebhookCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get cursor for slack webhook (%s)", err.Error())
		return webhooks, err
	}
	curCtx, curCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer curCancel()
	defer cursor.Close(curCtx)

	nxtCtx, nxtCancel := context.WithTimeout(wait.CtxSeconds(5))
	defer nxtCancel()
	for cursor.Next(nxtCtx) {
		webhook := v1.SlackWebhook{}
		if err := cursor.Decode(&webhook); err != nil {
			continue
		}
		webhooks = append(webhooks, webhook)
	}
	if cursor.Err() != nil {
		log.Errorf("failed to iterate slack webhook records (%s)", cursor.Err().Error())
	}

	return webhooks, nil
}

func updateSlackWebhookRecord(webhook v1.SlackWebhook) error {
	h := cubeMongo.GetGlobalHelper()
	filter := bson.M{"id": webhook.ID}
	update := bson.M{"$set": webhook}
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.SlackWebhookCollection(),
		filter,
		update,
	)
}

func deleteSlackWebhookRecord(id string) error {
	h := cubeMongo.GetGlobalHelper()
	return h.UpdateOne(
		v1.SettingsDB(),
		v1.SlackWebhookCollection(),
		bson.M{"id": id},
		bson.M{"$set": bson.M{"deleted": true}},
	)
}
