package tunings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (c *Controller) syncByDesiredAction(tuning definition.Tuning) error {
	switch tuning.Status.Desired {
	case status.Create, status.Update:
		return c.applyTuning(tuning)
	case status.Delete:
		return c.deleteTuning(tuning)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for tuning(%s)",
		tuning.Status.Desired,
		tuning.Name,
	)
}

func (c *Controller) deleteTuning(tuning definition.Tuning) error {
	policy, err := cubecos.GetPolicy()
	if err != nil {
		log.Errorf("failed to get all tunings: %s", err.Error())
		return err
	}

	policy.DeleteTuning(tuning.Name)
	err = cubecos.ApplyHexTunings(policy.Tunings)
	if err != nil {
		log.Errorf("failed to delete tunings: %s", err.Error())
		return err
	}

	err = cubecos.IsHexTuningDeleted(tuning)
	if err != nil {
		log.Errorf("failed to check if tuning %s is deleted: %s", tuning.Name, err.Error())
		return err
	}

	return nil
}

func (c *Controller) applyTuning(tuning definition.Tuning) error {
	policy, err := cubecos.GetPolicy()
	if err != nil {
		return err
	}

	policy.AppendTunings([]definition.Tuning{tuning})
	err = cubecos.ApplyHexTunings(policy.Tunings)
	if err != nil {
		log.Errorf("failed to apply tuning %s: %s", tuning.Name, err.Error())
		return err
	}

	err = cubecos.IsHexTuningApplied(tuning)
	if err != nil {
		log.Errorf("failed to check if tuning %s is applied: %s", tuning.Name, err.Error())
		return err
	}

	return nil
}
