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

	// err = h.parseAttributes()
	// if err != nil {
	// 	return nil, err
	// }

	// err = h.parseResponses()
	// if err != nil {
	// 	return nil, err
	// }

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

func (h *helper) delegateTriggerReq() error {
	return nil
}
