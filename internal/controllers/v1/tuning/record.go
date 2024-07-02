package tuning

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *Controller) updateTuningResult(tuning definition.Tuning) error {
	filter := bson.M{"node.id": tuning.Node.ID}
	tuning.Status.UpdatedAt = definition.TimeNowRFC3339()
	update := bson.M{"$set": tuning}

	return c.mongo.UpdateOne(
		definition.TuningDB(),
		definition.TuningCollection(tuning.Name),
		filter,
		update,
		mongo.CreateRecordIfNotExist,
	)
}

func (c *Controller) handleApplyingExit(tuning definition.Tuning, err error) {
	if err == nil {
		tuning.Status.Current = status.Completed
	} else {
		tuning.Status.Current = status.Error
		log.Errorf("failed to %s tuning %s: %s", tuning.Status.Desired, tuning.Name, err.Error())
	}

	tuning.Status.ClearDesired()
	err = c.updateTuningResult(tuning)
	if err != nil {
		log.Errorf("failed to update tuning result %s: %s", tuning.Name, err.Error())
	}
}

func (c *Controller) deleteTuningResult(tuning definition.Tuning) {
	filter := bson.M{"node.id": tuning.Node.ID, "name": tuning.Name}
	err := c.mongo.DeleteOne(
		definition.TuningDB(),
		definition.TuningCollection(tuning.Name),
		filter,
	)
	if err != nil {
		log.Errorf("failed to delete tuning result %s: %s", tuning.Name, err.Error())
	}
}

func (c *Controller) handleDeletionExit(tuning definition.Tuning, err error) {
	if err == nil {
		c.deleteTuningResult(tuning)
		return
	}

	tuning.Status.Current = status.Error
	log.Errorf("failed to %s tuning %s: %s", tuning.Status.Desired, tuning.Name, err.Error())
}

func (c *Controller) handleExit(tuning definition.Tuning, err error) {
	switch tuning.Status.Desired {
	case status.Create, status.Update:
		c.handleApplyingExit(tuning, err)
	case status.Delete:
		c.handleDeletionExit(tuning, err)
	}
}
