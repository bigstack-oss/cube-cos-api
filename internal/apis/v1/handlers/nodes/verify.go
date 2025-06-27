package nodes

import (
	"fmt"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) checkIpmiOperation() error {
	switch strings.ToLower(h.operation) {
	case "poweron", "poweroff", "powercycle":
		return nil
	default:
		return fmt.Errorf(
			"unsupport ipmi operation(%s), should be one of [poweron, poweroff, powercycle]",
			h.operation,
		)
	}
}

func (h *helper) checkStatusConflict() error {
	node, err := cubecos.GetNodeWithTimeSensitiveInfo(h.node)
	if err != nil {
		log.Errorf("nodes(%s): failed to get node(%v)", h.reqId, err)
		return err
	}

	switch h.operation {
	case "poweron":
		if node.Status == status.Up {
			return fmt.Errorf("node(%s) is already powered on", h.node)
		}
	case "poweroff":
		if node.Status == status.Down {
			return fmt.Errorf("node(%s) is already powered off", h.node)
		}
	case "powercycle":
		if node.Status != status.Up {
			return fmt.Errorf("node(%s) is not powered on, cannot power cycle", h.node)
		}
	}

	return nil
}
