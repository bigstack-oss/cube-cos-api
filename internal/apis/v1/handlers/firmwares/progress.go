package firmwares

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

type node struct {
	Name           string `json:"name"`
	nodes.Firmware `json:"firmware"`
}

func (h *helper) getUpgradeDetails() (*firmwares.Upgrade, error) {
	out, err := os.ReadFile(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read progress file(%v)", h.reqId, err)
		return nil, err
	}

	upgrade := &firmwares.Upgrade{}
	err = json.Unmarshal(out, upgrade)
	if err != nil {
		log.Errorf("firmwares(%s): failed to unmarshal progress file(%v)", h.reqId, err)
		return nil, err
	}

	return upgrade, nil
}

func (h *helper) sortUpgradeProgress(progresses *[]firmwares.Progress) {
	sort.Slice(*progresses, func(i, j int) bool {
		return (*progresses)[i].Host < (*progresses)[j].Host
	})
}

func (h *helper) syncFirstTimeInstallationProgress() {
	_, err := os.Stat(firmwares.UpdateProgress)
	if err == nil {
		return
	}

	if !os.IsNotExist(err) {
		log.Errorf("firmwares: failed to stat firmware progress file(%v)", err)
		return
	}

	f, err := os.Create(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares: failed to create firmware progress file(%v)", err)
		return
	}

	defer f.Close()
	version, err := cubecos.GetActiveFirmwareVersion()
	if err != nil {
		log.Errorf("firmwares: failed to get active firmware version(%v)", err)
		return
	}

	upgrade := firmwares.Upgrade{Version: version}
	for _, node := range nodes.List() {
		upgrade.Progresses = append(upgrade.Progresses, firmwares.Progress{
			Host: node.Hostname,
			Status: status.SystemUpdateProgress{
				Current:        status.Installed,
				ProcessPercent: 100,
			},
		})
	}

	b, err := json.Marshal(upgrade)
	if err != nil {
		log.Errorf("firmwares: failed to marshal firmware progress(%v)", err)
		return
	}

	_, err = f.Write(b)
	if err != nil {
		log.Errorf("firmwares: failed to write firmware progress(%v)", err)
		return
	}
}
