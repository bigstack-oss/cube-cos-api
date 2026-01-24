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
	err := h.removePreviousBoostrappingMarker()
	if err != nil {
		log.Errorf("firmwares(%s): failed to remove previous boostrapping marker (%v)", h.reqId, err)
		return err
	}

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

	h.delegateToPeers(updatables, &progress)
	cubecos.SetProgressDetails(&progress)
	go h.placeRollingTrigger()
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
	err := os.Remove(firmwares.UpdateProgress)
	if err != nil {
		log.Errorf("firmwares(%s): failed to remove update progress file (%v)", h.reqId, err)
		return err
	}

	h.syncFirstTimeInstallationProgress()
	h.syncProgressToControllers()
	return nil
}

func (h *helper) getFirmwareUpgradeProgress() (*firmwares.Upgrade, error) {
	h.syncFirstTimeInstallationProgress()
	upgrade, err := h.getUpgradeDetails()
	if err != nil {
		return nil, err
	}

	h.sortUpgradeProgress(&upgrade.Progresses)
	return upgrade, nil
}

func (h *helper) continueInterruptedFirmwareUpdate() error {
	node, err := nodes.Get(h.reqOpts.Hostname)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get node %s (%v)", h.reqId, h.reqOpts.Hostname, err)
		return err
	}

	err = cubecos.SetResolvedInfoBySsh(h.reqOpts.Hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to set resolved info on node %s (%v)", h.reqId, node.Hostname, err)
		return err
	}

	if node.IsVirtualIpOwner {
		cubecos.MoveVirtualIpOwner()
	}

	err = cubecos.SoftRebootBySsh(node.Hostname)
	if err != nil {
		log.Errorf("firmwares(%s): failed to soft reboot node %s (%v)", h.reqId, node.Hostname, err)
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
