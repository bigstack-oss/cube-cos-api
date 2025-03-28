package triggers

import (
	"errors"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	handler string
	trigger trigger.Options
	toggle  trigger.Toggle
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}

	switch handler {
	case "listTriggers", "getTrigger":
		return h, nil
	case "updateTrigger":
		return h.initUpdateHelper()
	case "enableOrDisableTrigger":
		return h.initToggleHelper()
	case "updateTriggerTask":
		return h.initTaskHelper()
	}

	return nil, errors.New("no internal function supported")
}

func (h *helper) initUpdateHelper() (*helper, error) {
	err := h.parseTrigger()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) initToggleHelper() (*helper, error) {
	name := h.c.Param("triggerName")
	if !cubecos.IsTriggerExist(name) {
		return nil, errors.New("trigger does not exist")
	}

	return h, nil
}

func (h *helper) initTaskHelper() (*helper, error) {
	err := h.parseTrigger()
	if err != nil {
		return nil, err
	}

	err = h.parseTaskId()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *helper) listTriggers() ([]trigger.Options, error) {
	triggers := []trigger.Options{}
	for _, trigger := range trigger.DefaultOptions {
		setResponseItemsToTrigger(&trigger)
		triggers = append(triggers, trigger)
	}

	return triggers, nil
}

func (h *helper) getTrigger(name string) (*trigger.Options, error) {
	for _, trigger := range trigger.DefaultOptions {
		if trigger.Name == name {
			setResponseItemsToTrigger(&trigger)
			return &trigger, nil
		}
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func (h *helper) delegateTriggerReq() {
	h.addReqRecord()
	reqQueue.Add(&h.trigger)
}

func (h *helper) checkTaskUpdateReq() error {
	if h.trigger.Id == "" {
		return fmt.Errorf("trigger id is required")
	}

	if h.trigger.Status == nil {
		return fmt.Errorf("trigger status is required")
	}

	return nil
}

func (h *helper) parseTriggerEnablement() error {
	err := h.c.ShouldBindJSON(&h.toggle)
	if err != nil {
		return err
	}

	h.trigger.Name = h.c.Param("triggerName")
	h.trigger.Enable = h.toggle.Enable
	h.trigger.InitUpdateStatus()
	return nil
}
