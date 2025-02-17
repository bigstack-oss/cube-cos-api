package settings

import (
	"context"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/google/uuid"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func upsertEmailSenderRecord(emailSender v1.EmailSender) error {
	h := cubeMongo.GetGlobalHelper()

	opts := options.Update().SetUpsert(true)
	if err := h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailSenderCollection(),
		bson.M{},
		bson.M{"$set": emailSender},
		opts,
	); err != nil {
		log.Errorf("failed to insert email sender record (%s)", err.Error())
		return err
	}
	return nil
}

func getEmailSenderRecord() (v1.EmailSender, error) {
	h := cubeMongo.GetGlobalHelper()
	emailSender := v1.EmailSender{}
	res, err := h.Get(v1.SettingsDB(), v1.EmailSenderCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get email sender record (%s)", err.Error())
		return emailSender, err
	}
	if err := res.Decode(&emailSender); err != nil {
		log.Errorf("failed to decode email sender record (%s)", err.Error())
		return emailSender, err
	}
	return emailSender, nil
}

func deleteEmailSenderRecord() error {
	h := cubeMongo.GetGlobalHelper()
	if err := h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailSenderCollection(),
		bson.M{},
		bson.M{"$set": bson.M{"deleted": true}},
	); err != nil {
		log.Errorf("failed to delete email sender record (%s)", err.Error())
		return err
	}
	return nil
}

func createEmailRecipientRecord(emailRecipient v1.EmailRecipient) error {
	h := cubeMongo.GetGlobalHelper()
	emailRecipient.ID = uuid.NewString()
	if err := h.Insert(
		v1.SettingsDB(),
		v1.EmailRecipientCollection(),
		emailRecipient,
	); err != nil {
		log.Errorf("failed to insert email recipient record (%s)", err.Error())
		return err
	}
	return nil
}

func getEmailRecipientRecords() ([]v1.EmailRecipient, error) {
	h := cubeMongo.GetGlobalHelper()
	recipients := []v1.EmailRecipient{}
	cursor, err := h.GetQueryCursor(v1.SettingsDB(), v1.EmailRecipientCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get cursor for email recipient (%s)", err.Error())
		return recipients, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		recipient := v1.EmailRecipient{}
		if err := cursor.Decode(&recipient); err != nil {
			log.Errorf("failed to decode email recipient record (%s)", err.Error())
			continue
		}
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}

func updateEmailRecipientRecord(emailRecipient v1.EmailRecipient) error {
	h := cubeMongo.GetGlobalHelper()
	filter := bson.M{"id": emailRecipient.ID}
	update := bson.M{"$set": emailRecipient}
	if err := h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailRecipientCollection(),
		filter,
		update,
	); err != nil {
		log.Errorf("failed to update email recipient record (%s)", err.Error())
		return err
	}
	return nil
}

func deleteEmailRecipientRecord(id string) error {
	h := cubeMongo.GetGlobalHelper()
	if err := h.UpdateOne(
		v1.SettingsDB(),
		v1.EmailRecipientCollection(),
		bson.M{"id": id},
		bson.M{"$set": bson.M{"deleted": true}},
	); err != nil {
		log.Errorf("failed to delete email recipient record (%s)", err.Error())
		return err
	}
	return nil
}

func createSlackWebhookRecord(webhook v1.SlackWebhook) error {
	h := cubeMongo.GetGlobalHelper()
	webhook.ID = uuid.NewString()
	if err := h.Insert(
		v1.SettingsDB(),
		v1.SlackWebhookCollection(),
		webhook,
	); err != nil {
		log.Errorf("failed to insert slack webhook record (%s)", err.Error())
		return err
	}
	return nil
}

func getSlackWebhookRecords() ([]v1.SlackWebhook, error) {
	h := cubeMongo.GetGlobalHelper()
	webhooks := []v1.SlackWebhook{}
	cursor, err := h.GetQueryCursor(v1.SettingsDB(), v1.SlackWebhookCollection(), bson.M{"deleted": bson.M{"$ne": true}})
	if err != nil {
		log.Errorf("failed to get cursor for slack webhook (%s)", err.Error())
		return webhooks, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		webhook := v1.SlackWebhook{}
		if err := cursor.Decode(&webhook); err != nil {
			log.Errorf("failed to decode slack webhook record (%s)", err.Error())
			continue
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, nil
}

func updateSlackWebhookRecord(webhook v1.SlackWebhook) error {
	h := cubeMongo.GetGlobalHelper()
	filter := bson.M{"id": webhook.ID}
	update := bson.M{"$set": webhook}
	if err := h.UpdateOne(
		v1.SettingsDB(),
		v1.SlackWebhookCollection(),
		filter,
		update,
	); err != nil {
		log.Errorf("failed to update slack webhook record (%s)", err.Error())
		return err
	}
	return nil
}
