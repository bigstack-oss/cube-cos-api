package triggers

import (
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c *gin.Context
}

func initReqHelper(c *gin.Context) (*helper, error) {
	return &helper{c: c}, nil
}

func (h *helper) getTriggers() ([]definition.Trigger, error) {
	return definition.DefaultTriggers, nil
}
