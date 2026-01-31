package firmwares

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	defssh "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ssh"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	cryptossh "golang.org/x/crypto/ssh"
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

func (h *helper) getUpgradeDetails() (*firmwares.Upgrade, error) {
	upgrade, err := cubecos.GetUpgradeProgress()
	if err != nil {
		return nil, err
	}

	if h.hasBootstrappingMarker() {
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
	for i, progress := range upgrade.Progresses {
		if progress.Phase != status.Rebooting && progress.Phase != status.Partitioning {
			continue
		}

		if i != 0 && upgrade.Progresses[i-1].Status.Current != status.Succeeded {
			continue
		}

		upgrade.Progresses[i].Phase = h.findPhaseFromBoostrapping(progress, boostrappings)
		upgrade.Progresses[i].Status.Current = h.findStatusFromBoostrapping(progress, boostrappings)
		if upgrade.Progresses[i].Status.Current == status.Succeeded || upgrade.Progresses[i].Status.Current == status.Resolved {
			upgrade.Progresses[i].Status.IsProcessing = false
			upgrade.Progresses[i].Status.ProcessPercent = 100
		} else {
			upgrade.Progresses[i].Status.IsProcessing = true
			upgrade.Progresses[i].Status.ProcessPercent = 80
		}
	}
}

func (h *helper) findPhaseFromBoostrapping(progress firmwares.Progress, boostrappings []firmwares.BoostrappingStatus) string {
	for _, boostrapping := range boostrappings {
		if boostrapping.Node != progress.Host {
			continue
		}

		if boostrapping.Return != "0" {
			return boostrapping.Stdout
		}
	}

	return status.Succeeded
}

func (h *helper) findStatusFromBoostrapping(progress firmwares.Progress, boostrappings []firmwares.BoostrappingStatus) string {
	for _, boostrapping := range boostrappings {
		if boostrapping.Node != progress.Host {
			continue
		}

		if boostrapping.Return != "0" {
			return status.Failed
		}

		if progress.Status.IsContinueAnywaied {
			return status.Resolved
		}

		return status.Succeeded
	}

	return status.Installing
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

	if h.IsClusterInBoostrapping() {
		h.syncByOtherNodes()
		return
	}

	h.setFreshFirmwareProgressFile()
}

func (h *helper) syncByOtherNodes() {
	for _, node := range nodes.List() {
		if node.IsLocal() {
			continue
		}

		if !h.doseNodeHasBootstrappingMarker(node) {
			continue
		}

		err := h.copyFirmwareDataFrom(node)
		if err != nil {
			log.Infof("firmwares: unable to copy firmware progress from node %s (%v)", node.Hostname, err)
			continue
		}

		return
	}
}

func (h *helper) setFreshFirmwareProgressFile() {
	file, err := os.Create(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares: failed to create firmware progress file(%v)", err)
		return
	}

	defer file.Close()
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

	_, err = file.Write(b)
	if err != nil {
		log.Errorf("firmwares: failed to write firmware progress(%v)", err)
	}
}

func (h *helper) syncProgressToAllNodes() {
	for _, nodes := range nodes.List() {
		if nodes.IsLocal() {
			continue
		}

		err := defssh.SyncRemoteFile(nodes.Hostname, firmwares.UpdateProgress, firmwares.UpdateProgress)
		if err != nil {
			log.Warnf("firmwares(%s): failed to sync firmware progress to controller %s(%v)", h.reqId, nodes.Hostname, err)
		}
	}
}

func (h *helper) getFinalInstallationStatus(node nodes.Node) string {
	if !node.IsLocal() {
		return h.getPeerInstallationStatus(node)
	}

	return status.Succeeded
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
		if h.isPendingInprogress(s) {
			(*list)[i].Status.IsProcessing = true
		}
	}
}

func (h *helper) isPendingInprogress(s string) bool {
	return s == status.Installing || s == status.WaitingReboot || s == status.Rebooting
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

func (h *helper) syncNodeUpgradeProgress(hostname string, upgrade *firmwares.Upgrade, s *status.SystemUpdateProgress) {
	for i, p := range upgrade.Progresses {
		if p.Host != hostname {
			continue
		}

		upgrade.Progresses[i].Status = *s
		return
	}

	upgrade.Progresses = append(
		upgrade.Progresses,
		firmwares.Progress{
			Host:   hostname,
			Phase:  status.Partitioning,
			Status: *s,
		},
	)
}

func (h *helper) waitForPrimaryControllerVmEvacuated() {
	hostname, err := cubecos.GetPrimaryControllerHost()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get primary controller host(%v)", err, h.reqId)
		h.markNodeAsFailed(err.Error())
		return
	}

	log.Infof("firmwares: wait for all vms evacuated on node %s", hostname)
	err = cubecos.WaitForAllVmsEvacuated(hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to wait for all vms evacuated (%v)", err, h.reqId)
		h.markNodeAsFailed(err.Error())
		return
	}

	err = cubecos.SetNodeUpdateProgress(hostname, status.Rebooting, status.Rebooting)
	if err != nil {
		log.Errorf("firmwares(%s): failed to set rebooting progress(%v)", err, h.reqId)
		h.markNodeAsFailed(err.Error())
		return
	}

	cubecos.SyncFirmwareUpgradeProgressToAllNodes()
}

func (h *helper) IsClusterInBoostrapping() bool {
	for _, node := range nodes.List() {
		if h.doseNodeHasBootstrappingMarker(node) {
			return true
		}
	}

	return false
}

func (h *helper) doseNodeHasBootstrappingMarker(node nodes.Node) bool {
	resp, err := h.http.R().
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.GetFirmwareBootstrappingUrl())
	if err != nil {
		log.Infof("firmwares: unable to find firmware bootstrapping marker from node %s (%v)", node.Hostname, err)
		return false
	}

	return resp.IsError()
}

func (h *helper) copyFirmwareDataFrom(node nodes.Node) error {
	sshAuth, err := defssh.GenSshAuth(defssh.DefaultPrivateKey)
	if err != nil {
		return err
	}

	ssh, err := ssh.NewHelper(
		ssh.Host(fmt.Sprintf("%s:22", node.Hostname)),
		ssh.User("root"),
		ssh.AuthMethod(sshAuth),
		ssh.HostKeyCallback(cryptossh.InsecureIgnoreHostKey()),
	)
	if err != nil {
		return err
	}

	defer ssh.Close()
	err = ssh.CopyFrom(firmwares.BoostrappingMarker, firmwares.BoostrappingMarker)
	if err != nil {
		log.Errorf("firmwares(%s): failed to copy boostrapping marker to controller %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	err = ssh.CopyFrom(firmwares.UpdateProgress, firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares(%s): failed to copy firmware upgrade progress from node %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	err = ssh.CopyFrom(firmwares.ResolvedMarker, firmwares.ResolvedMarker)
	if err != nil {
		log.Errorf("firmwares(%s): failed to copy resolved marker from node %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	return nil
}
