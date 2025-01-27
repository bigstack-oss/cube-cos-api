package healths

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/errors"
)

func checkRepairCondition() error {
	if cubecos.IsRepairing() {
		return errors.DataCenterIsRepairing
	}

	if !cubecos.IsRepairable() {
		return errors.DataCenterIsNotReady
	}

	return nil
}
