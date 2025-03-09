package triggers

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) operateReq(trigger trigger.Options) error {
	switch trigger.Status.Desired {
	case status.Update:
		return o.updateTrigger(trigger)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for trigger(%s)",
		trigger.Status.Desired,
		trigger.Name,
	)
}

func (o *Operator) updateTrigger(trigger trigger.Options) error {
	policy, err := cubecos.GetTriggerPolicy()
	if err != nil {
		return err
	}

	policy.UpdateOrAppendTrigger(trigger)
	err = cubecos.ApplyTriggers(policy.Triggers)
	if err != nil {
		log.Errorf("failed to apply trigger %s: %s", trigger.Name, err.Error())
		return err
	}

	err = cubecos.IsTriggerApplied(trigger)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	return nil
}
