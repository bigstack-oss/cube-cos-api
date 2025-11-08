package firmwares

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

type node struct {
	Name           string `json:"name"`
	nodes.Firmware `json:"firmware"`
}

func (h *helper) initUpgradeProgress() firmwares.Upgrade {
	return firmwares.Upgrade{
		Version:          h.reqOpts.Version,
		IsRollingApplied: h.reqOpts.AutoRolling,
		Progresses: []firmwares.Progress{
			{
				Host:  base.Hostname,
				Phase: status.Partitioning,
				Status: status.SystemUpdateProgress{
					Current:        status.Installing,
					IsProcessing:   true,
					ProcessPercent: 30,
				},
			},
		},
	}
}

func (h *helper) setProgressDetails(progress firmwares.Upgrade) {
	file, err := os.Create(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create progress file for update(%v)", h.reqId, err)
		return
	}

	defer file.Close()
	content, err := json.Marshal(progress)
	if err != nil {
		log.Errorf("firmwares(%s): failed to marshal progress details(%v)", h.reqId, err)
		return
	}

	_, err = file.WriteString(string(content))
	if err != nil {
		log.Errorf("firmwares(%s): failed to write progress file(%v)", h.reqId, err)
		return
	}
}

func (h *helper) getUpgradeDetails() (*firmwares.Upgrade, error) {
	upgrade, err := h.getPartitioningProgress()
	if err != nil {
		return nil, err
	}

	if h.isBoostrappingInProgress() {
		return h.syncBoostrappingProgress(upgrade)
	}

	return upgrade, nil
}

func (h *helper) syncBoostrappingProgress(upgrade *firmwares.Upgrade) (*firmwares.Upgrade, error) {
	boostrapping, err := cubecos.GetBoostrappingProgress()
	if err != nil {
		return nil, err
	}

	h.convertToUpgradeProgress(upgrade, boostrapping)
	return upgrade, nil
}

func (h *helper) convertToUpgradeProgress(upgrade *firmwares.Upgrade, boostrappings []firmwares.BoostrappingStatus) {
	progresses := []firmwares.Progress{}

	for _, boostrapping := range boostrappings {
		progress := firmwares.Progress{
			Host:  boostrapping.Node,
			Phase: boostrapping.Stdout,
			Status: status.SystemUpdateProgress{
				Current:        h.convertProgressStatus(boostrapping),
				IsProcessing:   true,
				ProcessPercent: 80,
			},
		}

		progresses = append(progresses, progress)
	}

	upgrade.Progresses = progresses
}

func (h *helper) convertProgressStatus(bootstrapping firmwares.BoostrappingStatus) string {
	if bootstrapping.Return != "0" {
		return status.Failed
	}

	if strings.Contains(bootstrapping.Stdout, "succeeded") {
		return status.Succeeded
	}

	return status.Installing
}

func (h *helper) getPartitioningProgress() (*firmwares.Upgrade, error) {
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

	upgrade := firmwares.Upgrade{Version: version, IsRollingApplied: h.reqOpts.AutoRolling}
	for _, node := range nodes.List() {
		upgrade.Progresses = append(upgrade.Progresses, firmwares.Progress{
			Host: node.Hostname,
			Status: status.SystemUpdateProgress{
				Current:        h.getFinalInstallationStatus(node),
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

func (h *helper) syncProgressToControllers() error {
	controllers, err := nodes.GetPeerControls()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get peer controllers (%v)", h.reqId, err)
		return err
	}

	for _, controller := range controllers {
		err := ssh.SyncRemoteFile(controller.Hostname, firmwares.UpdateProgress, firmwares.UpdateProgress)
		if err != nil {
			log.Warnf("firmwares(%s): failed to sync firmware progress to controller %s(%v)", h.reqId, controller.Hostname, err)
		}
	}

	return nil
}

func (h *helper) getFinalInstallationStatus(node nodes.Node) string {
	if !node.IsLocal() {
		return h.getPeerInstallationStatus(node)
	}

	_, err := os.Stat(firmwares.ResolvedMarker)
	if err != nil {
		return status.Succeeded
	}

	return status.Resolved
}

func (h *helper) getPeerInstallationStatus(node nodes.Node) string {
	req := h.http.R().
		SetResult(&firmwares.ResolvedStatus{}).
		SetHeaders(h.convertHeadersToMap(h.c.Request.Header))
	resp, err := req.Execute(h.c.Request.Method, node.GetFirmwareResovledUrl())
	if err != nil {
		log.Errorf("firmwares(%s): failed to get peer resolved info %s(%v)", h.reqId, node.Hostname, err)
		return status.Succeeded
	}

	if resp.IsError() {
		log.Errorf("firmwares(%s): has resp error from peer %s(%s)", h.reqId, node.Hostname, resp.String())
		return status.Succeeded
	}

	if !resp.Result().(*firmwares.ResolvedStatus).HasFailureBeenResolved {
		return status.Succeeded
	}

	return status.Resolved
}

func (h *helper) syncFirmwareStatuses(list *[]firmwares.Firmware) {
	upgrade, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get firmware upgrade progress(%v)", h.reqId, err)
		return
	}

	for i, firmware := range *list {
		if firmware.Version != upgrade.Version {
			continue
		}

		s := h.syncOverallProgressStatus(upgrade.Progresses)
		(*list)[i].Status.Current = s
		if s == status.Installing {
			(*list)[i].Status.IsProcessing = true
		}
	}
}

func (h *helper) syncOverallProgressStatus(progresses []firmwares.Progress) string {
	statusMap := map[string]bool{}
	for _, progress := range progresses {
		statusMap[progress.Status.Current] = true
	}

	if statusMap[status.Installing] {
		return status.Installing
	}

	if statusMap[status.WaitingReboot] {
		return status.WaitingReboot
	}

	if statusMap[status.Rebooting] {
		return status.Rebooting
	}

	if statusMap[status.Failed] {
		return status.Failed
	}

	return status.Succeeded
}
