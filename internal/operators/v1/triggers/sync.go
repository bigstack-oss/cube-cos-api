package triggers

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
)

func (o *Operator) operateReq(trigger trigger.ApiOptions) error {
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

func (o *Operator) updateTrigger(trigger trigger.ApiOptions) error {
	cosTrigger := trigger.ConvertToCosOptions()
	return cubecos.ApplyTrigger(cosTrigger)
}
