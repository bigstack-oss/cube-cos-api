package v1

import (
	"strings"

	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbTunings = "tunings"
)

var (
	CreateRecordIfNotExist = options.Update().SetUpsert(true)
)

func GenCollectionNameByTuningName(name string) string {
	return strings.Split(name, ".")[0]
}

func TuningDB() string {
	return dbTunings
}

func TuningCollection(name string) string {
	return strings.Split(name, ".")[0]
}
