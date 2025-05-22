package tunings

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/tunings"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateReq(tuning tunings.Tuning) error {
	switch tuning.Status.Desired {
	case status.Updated:
		return o.updateTuning(tuning)
	case status.Reset:
		return o.resetTuning(tuning)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for tuning(%s)",
		tuning.Status.Desired,
		tuning.Name,
	)
}

func (o *Operator) resetTuning(tuning tunings.Tuning) error {
	policy, err := cubecos.GetTuningPolicy(cubecos.TuningPolicyFile)
	if err != nil {
		log.Errorf("tuning: failed to get all tunings(%v)", err)
		return err
	}

	policy.DeleteTuning(tuning.Name)
	err = cubecos.ApplyTunings(policy.Tunings)
	if err != nil {
		log.Errorf("tuning: failed to delete %s: %v", tuning.Name, err)
		return err
	}

	if !cubecos.IsTuningDeleted(tuning) {
		err := fmt.Errorf("tuning: %s is not reset", tuning.Name)
		log.Errorf(err.Error())
		return err
	}

	return nil
}

func (o *Operator) updateTuning(tuning tunings.Tuning) error {
	policy, err := cubecos.GetTuningPolicy(cubecos.TuningPolicyFile)
	if err != nil {
		return err
	}

	policy.UpdateOrAppendTuning(tuning)
	err = cubecos.ApplyTunings(policy.Tunings)
	if err != nil {
		log.Errorf("failed to apply tuning %s: %v", tuning.Name, err)
		return err
	}

	err = cubecos.IsTuningApplied(tuning)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}
