package triggers

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
)

func (o *Operator) operateReq(req triggers.ReqOpts) error {
	switch req.Status.Desired {
	case status.Updated:
		return o.updateTrigger(req)
	case status.Deleted:
		return o.deleteTrigger(req)
	}

	return fmt.Errorf(
		"unknown desired action(%s) for trigger(%s)",
		req.Status.Desired,
		req.Name,
	)
}

func (o *Operator) updateTrigger(req triggers.ReqOpts) error {
	// cosTrigger := trigger.ToCosSchema()

	err := o.syncScripts(req.ReqResponse.Script)
	if err != nil {
		return err
	}

	trigger := o.convertToTrigger(req)
	return cubecos.ApplyTrigger(trigger)
}

func (o *Operator) deleteTrigger(req triggers.ReqOpts) error {
	return cubecos.DeleteTrigger(
		triggers.Trigger{Name: req.Name},
	)
}

func (o *Operator) convertToTrigger(req triggers.ReqOpts) triggers.Trigger {
	return triggers.Trigger{
		Name:        req.Name,
		Enabled:     req.Enabled,
		Description: req.Description,
		Match:       req.GenMatchRule(),
		Responses: triggers.Responses{
			Emails: req.ReqResponse.Emails,
			Slacks: req.ReqResponse.Slacks,
			Execs: triggers.Execs{
				Shells: []string{req.ReqResponse.Script.Name + ".shell"},
			},
		},
	}
}
