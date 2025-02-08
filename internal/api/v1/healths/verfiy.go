package healths

import (
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/errors"
)

// M1 TODO:
// The cubecos.IsRepairing() will be relied on the /var/run/{markerfile} from COS SDk
// COS dev is working on the implementation, once it's ready, the logic of IsRepairing() will be updated.
func checkEnvCondition() error {
	if cubecos.IsRepairing() {
		return errors.DataCenterIsRepairing
	}

	if !cubecos.IsRepairable() {
		return errors.DataCenterIsNotReady
	}

	return nil
}
