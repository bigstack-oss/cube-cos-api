package firmwares

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

type node struct {
	Name           string `json:"name"`
	nodes.Firmware `json:"firmware"`
}

type upgrade struct {
	Version    string     `json:"version"`
	Progresses []progress `json:"progresses"`
}

type progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

func (h *helper) hasInprogressUpdate() bool {
	_, err := os.Stat(firmwares.UpdateProgress)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	log.Errorf(
		"firmwares(%s): failed to stat progress file(%v)",
		h.reqId, err,
	)

	return false
}

func (h *helper) getUpgradeDetails() (*upgrade, error) {
	out, err := os.ReadFile(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read progress file(%v)", h.reqId, err)
		return nil, err
	}

	upgrade := &upgrade{}
	err = json.Unmarshal(out, upgrade)
	if err != nil {
		log.Errorf("firmwares(%s): failed to unmarshal progress file(%v)", h.reqId, err)
		return nil, err
	}

	return upgrade, nil
}

func (h *helper) sortUpgradeProgress(progresses *[]progress) {
	sort.Slice(*progresses, func(i, j int) bool {
		return (*progresses)[i].Host < (*progresses)[j].Host
	})
}
