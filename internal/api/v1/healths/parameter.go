package healths

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
)

func parseModule(c *gin.Context) (*definition.Module, error) {
	m := c.Param("moduleType")
	module, found := cubecos.Modules[m]
	if !found {
		return nil, fmt.Errorf("module(%s) not found", m)
	}

	return &module, nil
}
