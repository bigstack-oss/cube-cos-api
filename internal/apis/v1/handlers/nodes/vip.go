package nodes

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pacemaker"
	log "go-micro.dev/v5/logger"
)

func (h *helper) waitForVirutalIpOwnerChanged(oldOwner string) error {
	for range 600 {
		wait.Seconds(1)
		host, err := pacemaker.GetVirtualIpHost()
		if err != nil {
			log.Errorf("nodes(%s): failed to get virtual ip host(%v)", h.reqId, err)
			continue
		}

		if host == oldOwner {
			log.Infof("nodes(%s): virtual ip owner is still %s, wait for it changed", h.reqId, oldOwner)
			continue
		}

		return nil
	}

	return fmt.Errorf(
		"failed to wait for virtual ip owner changed in 10 minutes",
	)
}
