package healths

import (
	"fmt"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func (h *helper) parseModule() (*v1.Module, error) {
	m := h.c.Param("moduleType")
	module, found := cubecos.Modules[m]
	if !found {
		return nil, fmt.Errorf("module(%s) not found", m)
	}

	return &module, nil
}
