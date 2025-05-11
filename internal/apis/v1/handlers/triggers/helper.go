package triggers

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	http http.Helper

	trigger               triggers.ApiSchema
	toggle                triggers.Toggle
	rawBody               []byte
	isClusterWiseRequired bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
		http:    *http.GetGlobalHelper(),
		rawBody: bodies.ParseReq(c),
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listTriggers() ([]triggers.ApiSchema, error) {
	list := []triggers.ApiSchema{}
	for _, trigger := range triggers.List() {
		h.syncUpdatingInfo(&trigger)
		list = append(list, trigger)
	}

	return list, nil
}

func (h *helper) getTrigger(name string) (*triggers.ApiSchema, error) {
	for _, trigger := range triggers.List() {
		if trigger.Name == name {
			return &trigger, nil
		}
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func (h *helper) checkTaskUpdateReq() error {
	if h.trigger.Name == "" {
		return fmt.Errorf("trigger id is required")
	}

	if h.trigger.Status == nil {
		return fmt.Errorf("trigger status is required")
	}

	return nil
}
