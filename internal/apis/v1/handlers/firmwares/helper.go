package firmwares

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	opfirmwares "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/firmwares"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = opfirmwares.ReqQueue
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string
	http    *http.Helper
	mongo   *mongo.Helper

	file    string
	reqOpts firmwares.ReqOpts
	page    *pages.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		http:    http.GetGlobalHelper(),
		mongo:   mongo.GetGlobalHelper(),
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listFirmwares() (*firmwarePage, error) {
	firmwares, err := cubecos.ListFirmwares()
	if err != nil {
		log.Errorf("firmwares(%s): failed to list firmwares(%v)", h.reqId, err)
		return nil, err
	}

	h.syncFirmwareStatuses(&firmwares)
	h.sortFirmwares(&firmwares)
	return &firmwarePage{
		Firmwares: h.paginateFirmwares(firmwares),
		Page:      h.genPageInfo(firmwares),
	}, nil
}

func (h *helper) listUpdatableNodes() ([]node, error) {
	list := nodes.List()
	if len(list) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	updatables := make([]node, 0, len(list))
	for _, n := range list {
		updatables = append(updatables, node{
			Name: n.Hostname,
			Firmware: nodes.Firmware{
				Active:   n.Firmware.Active,
				Inactive: n.Firmware.Inactive,
			},
		})
	}

	h.sortNodesByName(&updatables)
	return updatables, nil
}

func (h *helper) sortNodesByName(nodes *[]node) {
	sort.SliceStable(*nodes, func(i, j int) bool {
		return (*nodes)[i].Name < (*nodes)[j].Name
	})
}

func (h *helper) updateFirmware() error {
	if h.reqOpts.AutoRolling {
		return h.startRollingUpdate()
	}

	h.resetBootstrappingLogs()
	h.delegateToLocal()
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return nil
	}

	progress := h.initUpgradeProgress()
	updatables, err := h.listUpdatableNodes()
	if err != nil {
		log.Errorf("firmwares(%s): failed to list updatable nodes (%v)", h.reqId, err)
		return err
	}

	h.updatePeerFirmware(updatables, &progress)
	cubecos.SetProgressDetails(&progress)
	h.syncProgressToAllNodes()
	return nil
}

// The roll is one cluster-wide job owned by the hex_sdk state machine: it
// stages the package on every node, refuses to reboot anything unless all
// nodes staged, then rolls node by node. The API only starts it and polls.
func (h *helper) startRollingUpdate() error {
	roll, err := cubecos.GetRollStatus()
	if err != nil {
		return err
	}

	if roll.IsInFlight() {
		return fmt.Errorf("a rolling update is already in progress (%s)", roll.State)
	}

	h.resetBootstrappingLogs()
	log.Infof("firmwares(%s): starting rolling update with %s", h.reqId, h.reqOpts.PkgPath)
	go func(reqId string, pkg string) {
		err := cubecos.StartRollingUpdate(pkg)
		if err != nil {
			log.Errorf("firmwares(%s): rolling update %s ended with an error(%v)", reqId, pkg, err)
		}
	}(h.reqId, h.reqOpts.PkgPath)

	return nil
}

func (h *helper) updateNodeFirmware() error {
	update, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get firmware upgrade progress (%v)", h.reqId, err)
		return err
	}

	defer cubecos.SetProgressDetails(update)
	for i, p := range update.Progresses {
		if p.Host != h.reqOpts.Hostname {
			continue
		}

		update.Progresses[i].Status.Current = h.reqOpts.Status.Current
		update.Progresses[i].Status.ProcessPercent = 30
		update.Progresses[i].Status.IsProcessing = true
		update.Progresses[i].Status.Description = ""
		break
	}

	if nodes.IsLocal(h.reqOpts.Hostname) {
		h.delegateToLocal()
	}

	return h.delegateToPeer(h.reqOpts.Hostname, update)
}

func (h *helper) abortFirmwareUpdate() error {
	err := h.abortRoll()
	if err != nil {
		return err
	}

	h.removeClusterFirmwareUpgradeProgress()
	h.syncFirstTimeInstallationProgress()
	h.syncProgressToAllNodes()
	return nil
}

// Only an in-flight upgrade roll is aborted: a rolling restart is not a
// firmware update and must not be stopped from here, and skipping when there is
// nothing to abort keeps the endpoint idempotent.
func (h *helper) abortRoll() error {
	roll, err := cubecos.GetRoll()
	if err != nil {
		// Preserve the previous behaviour: still clear the local record.
		log.Errorf("firmwares(%s): failed to read roll job to abort(%v)", h.reqId, err)
		return nil
	}

	if !roll.IsUpgrade() || !roll.IsInFlight() {
		return nil
	}

	log.Infof("firmwares(%s): aborting the in-flight rolling update", h.reqId)
	return cubecos.AbortRoll()
}

func (h *helper) getFirmwareUpgradeProgress() (*firmwares.Upgrade, error) {
	roll, err := cubecos.GetRoll()
	if err != nil {
		return nil, err
	}

	if roll.IsUpgrade() && roll.IsInFlight() {
		return h.convertRollToUpgrade(roll)
	}

	// No firmware roll in flight: report the cluster as it stands from the
	// local record.
	h.syncFirstTimeInstallationProgress()
	upgrade, err := h.getUpgradeDetails()
	if err != nil {
		return nil, err
	}

	h.sortUpgradeProgress(&upgrade.Progresses)
	return upgrade, nil
}

func (h *helper) convertRollToUpgrade(roll *firmwares.Roll) (*firmwares.Upgrade, error) {
	rollStatus, err := cubecos.GetRollStatus()
	if err != nil {
		return nil, err
	}

	upgrade := &firmwares.Upgrade{
		Version:          h.getRollTargetVersion(roll),
		IsRollingApplied: rollStatus.IsRollingApplied,
		Progresses:       rollStatus.Progresses,
	}

	h.sortUpgradeProgress(&upgrade.Progresses)
	return upgrade, nil
}

// The roll job records the target package, not the display version the UI
// matches firmware list entries on, so convert it back.
func (h *helper) getRollTargetVersion(roll *firmwares.Roll) string {
	if roll.Version == "" {
		return base.ActiveFirmwareVersion
	}

	firmware, err := cubecos.ConvertPkgNameToFirmware(filepath.Base(roll.Version))
	if err != nil {
		log.Errorf("firmwares(%s): failed to convert roll package %s(%v)", h.reqId, roll.Version, err)
		return base.ActiveFirmwareVersion
	}

	return firmware.Version
}

func (h *helper) continueInterruptedFirmwareUpdate() error {
	node, err := nodes.Get(h.reqOpts.Hostname)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get node %s (%v)", h.reqId, h.reqOpts.Hostname, err)
		return err
	}

	err = cubecos.SetNodeUpdateProgress(node.Hostname, status.Rebooting, status.Rebooting)
	if err != nil {
		log.Errorf("firmwares(%s): failed to set rebooting progress(%v)", err, h.reqId)
		return err
	}

	err = cubecos.SetNodeAsContinueAnywaied(node.Hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to set continue anywaied flag(%v)", h.reqId, err)
		return err
	}

	cubecos.SyncFirmwareUpgradeProgressToAllNodes()
	if node.IsVirtualIpOwner {
		cubecos.MoveVirtualIpOwner()
	}

	err = cubecos.SoftRebootBySsh(node.Hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to soft reboot node %s(%v)", h.reqId, node.Hostname, err)
		return err
	}

	return nil
}

func (h *helper) deleteFirmware() error {
	err := h.checkFirmwarePattern()
	if err != nil {
		return err
	}

	segments := strings.Split(h.file, " ")
	version := segments[3]
	entries, err := os.ReadDir(firmwares.UpdateDir)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read update directory %s(%v)", h.reqId, firmwares.UpdateDir, err)
		return errors.New("has an error during reading internal fs space")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(firmwares.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".pkg") {
			continue
		}

		if !strings.Contains(file, version) {
			continue
		}

		err = os.Remove(file)
		if err == nil {
			return nil
		}

		log.Errorf("firmwares(%s): failed to delete firmware file %s (%v)", h.reqId, file, err)
		return err
	}

	return fmt.Errorf(
		"firmware version %s not found",
		version,
	)
}
