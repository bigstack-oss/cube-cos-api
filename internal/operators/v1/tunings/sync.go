package tunings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateReq(tuning definition.Tuning) error {
	switch tuning.Status.Desired {
	case status.Update:
		return o.applyTuning(tuning)
	case status.Reset:
		return o.deleteTuning(tuning)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for tuning(%s)",
		tuning.Status.Desired,
		tuning.Name,
	)
}

func (o *Operator) deleteTuning(tuning definition.Tuning) error {
	policy, err := cubecos.GetTuningPolicy(cubecos.TuningPolicyFile)
	if err != nil {
		log.Errorf("failed to get all tunings: %s", err.Error())
		return err
	}

	policy.DeleteTuning(tuning.Name)
	err = cubecos.ApplyTunings(policy.Tunings)
	if err != nil {
		log.Errorf("failed to delete tunings: %s", err.Error())
		return err
	}

	if !cubecos.IsTuningDeleted(tuning) {
		err := fmt.Errorf("tuning %s is not deleted", tuning.Name)
		log.Errorf(err.Error())
		return err
	}

	return nil
}

func (o *Operator) applyTuning(tuning definition.Tuning) error {
	policy, err := cubecos.GetTuningPolicy(cubecos.TuningPolicyFile)
	if err != nil {
		return err
	}

	policy.AppendTunings([]definition.Tuning{tuning})
	err = cubecos.ApplyTunings(policy.Tunings)
	if err != nil {
		log.Errorf("failed to apply tuning %s: %s", tuning.Name, err.Error())
		return err
	}

	err = cubecos.IsTuningApplied(tuning)
	if err != nil {
		log.Errorf("failed to check if tuning %s is applied: %s", tuning.Name, err.Error())
		return err
	}

	return nil
}
