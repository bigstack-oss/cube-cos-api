package triggers

import (
	"errors"
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/trigger"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	handler string

	trigger trigger.Options
}

func initReqHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}

	switch handler {
	case "listTriggers", "getTrigger":
		return h, nil
	case "updateTrigger":
		return h.initUpdateHelper()
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
