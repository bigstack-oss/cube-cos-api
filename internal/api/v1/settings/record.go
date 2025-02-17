package settings

import (
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
