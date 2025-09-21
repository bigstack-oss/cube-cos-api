package firmwares

import (
	"sync/atomic"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

var (
	rollingTriggerCount = int32(0)
)

func (h *helper) placeRollingTrigger() {
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return
	}

	if !h.reqOpts.AutoRolling {
		return
	}

	if atomic.LoadInt32(&rollingTriggerCount) > 0 {
		return
	}

	atomic.AddInt32(&rollingTriggerCount, 1)
	defer atomic.AddInt32(&rollingTriggerCount, -1)
	for {
		wait.Seconds(5)
		if !h.areAllNodesPartitioned() {
			continue
		}

		err := cubecos.EvacuateVms(base.Hostname)
		if err != nil {
			h.markNodeAsFailed(err.Error())
			return
		}

		err = cubecos.GracefulReboot()
		if err != nil {
			h.markNodeAsFailed(err.Error())
			return
		}
	}
}

func (h *helper) areAllNodesPartitioned() bool {
	update, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		return false
	}

	for _, progress := range update.Progresses {
		if progress.Status.Current != "updated" {
			return false
		}
	}

	return true
}

func (h *helper) markNodeAsFailed(errMsg string) {
	update, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		return
	}

	for i, progress := range update.Progresses {
		if progress.Host != base.Hostname {
			continue
		}

		update.Progresses[i].Status.Current = "failed"
		update.Progresses[i].Status.Description = errMsg
		break
	}

	h.setProgressDetails(*update)
}
