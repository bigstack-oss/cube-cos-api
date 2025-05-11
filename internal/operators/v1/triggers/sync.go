package triggers

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (o *Operator) operateReq(trigger triggers.ApiSchema) error {
	switch trigger.Status.Desired {
	case status.Updated:
		return o.updateTrigger(trigger)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for trigger(%s)",
		trigger.Status.Desired,
		trigger.Name,
	)
}

func (o *Operator) updateTrigger(trigger triggers.ApiSchema) error {
	cosTrigger := trigger.ToCosSchema()
	return cubecos.ApplyTrigger(cosTrigger)
}
