package tunings

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (o *Operator) updateTuningResult(tuning definition.Tuning) error {
	// filter := bson.M{"node.id": tuning.Node.Id}
	// tuning.Status.UpdatedAt = definition.TimeNowRFC3339()
	// update := bson.M{"$set": tuning}

	// return o.mongo.UpdateOne(
	// 	definition.TuningDB(),
	// 	definition.TuningCollection(tuning.Name),
	// 	filter,
	// 	update,
	// 	mongo.CreateRecordIfNotExist,
	// )

	return nil
}

func (o *Operator) handleApplyingExit(tuning definition.Tuning, err error) {
	// if err == nil {
	// 	tuning.Status.Current = status.Completed
	// } else {
	// 	tuning.Status.Current = status.Error
	// 	log.Errorf("failed to %s tuning %s: %s", tuning.Status.Desired, tuning.Name, err.Error())
	// }

	// tuning.Status.ClearDesired()
	// err = o.updateTuningResult(tuning)
	// if err != nil {
	// 	log.Errorf("failed to update tuning result %s: %s", tuning.Name, err.Error())
	// }
}

func (o *Operator) deleteTuningResult(tuning definition.Tuning) {
	filter := bson.M{"node.id": tuning.Node.Id, "name": tuning.Name}
	err := o.mongo.DeleteOne(
		definition.TuningDB(),
		definition.TuningCollection(tuning.Name),
		filter,
	)
	if err != nil {
		log.Errorf("failed to delete tuning result %s: %s", tuning.Name, err.Error())
	}
}

func (o *Operator) handleDeletionExit(tuning definition.Tuning, err error) {
	// if err == nil {
	// 	o.deleteTuningResult(tuning)
	// 	return
	// }

	// tuning.Status.Current = status.Error
	// log.Errorf("failed to %s tuning %s: %s", tuning.Status.Desired, tuning.Name, err.Error())
}

func (o *Operator) handleExit(tuning definition.Tuning, err error) {
	// switch tuning.Status.Desired {
	// case status.Create, status.Update:
	// 	o.handleApplyingExit(tuning, err)
	// case status.Delete:
	// 	o.handleDeletionExit(tuning, err)
	// }
}
