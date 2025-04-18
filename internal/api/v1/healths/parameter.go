package healths

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

const (
	defaultPastOneHour = "24h"
)

func parseModule(c *gin.Context) (*v1.Module, error) {
	m := c.Param("moduleType")
	module, found := cubecos.Modules[m]
	if !found {
		return nil, fmt.Errorf("module(%s) not found", m)
	}

	return &module, nil
}
