package healths

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
)

func (h *helper) checkEnvCondition() error {
	if cubecos.IsRepairing() {
		return errors.ErrDataCenterIsRepairing
	}

	if !cubecos.IsRepairable() {
		return errors.ErrDataCenterIsNotReady
	}

	return nil
}
