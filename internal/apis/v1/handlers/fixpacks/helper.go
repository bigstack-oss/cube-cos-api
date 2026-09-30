package fixpacks

import (
	"fmt"
	"os"
	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	http  *http.Helper
	mongo *mongo.Helper

	file    string
	reqOpts fixpacks.ReqOpts
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

func (h *helper) listFixpacks(syncNodeProgress bool) (*fixpacksPage, error) {
	fixpacks, err := cubecos.ListFixpacks()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to list fixpacks(%v)", h.reqId, err)
		return nil, err
	}

	h.syncRequestingRecord(&fixpacks)
	if syncNodeProgress {
		h.syncStatusByNodeProgresses(&fixpacks)
	}

	return &fixpacksPage{
		Fixpacks: h.paginateFixpacks(fixpacks),
		Page:     h.genPageInfo(fixpacks),
	}, nil
}

func (h *helper) getFixpackUpdateProgress(version string) (*update, error) {
	update, err := h.getUpgradeProgressRecordByVersion(version)
	if err != nil {
		return nil, err
	}

	h.sortUpdateProgress(&update.Progresses)
	return update, nil
}

func (h *helper) listUpdatableNodes(version string) ([]node, error) {
	list := nodes.List()
	if len(list) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	updatables := h.convertToUpdatableNodes(list)
	updatables, err := h.filterUnsupportedNodes(updatables, version)
	if err != nil {
		return nil, err
	}

	h.markInstalledNodes(updatables, version)
	h.sortNodesByHost(&updatables)
	return updatables, nil
}

// markInstalledNodes flags the nodes whose own history has version installed.
func (h *helper) markInstalledNodes(list []node, version string) {
	statuses, err := cubecos.ListFixpackNodeStatus(version)
	if err != nil {
		log.Warnf("fixpacks(%s): failed to get node fixpack status(%v)", h.reqId, err)
		return
	}

	for i := range list {
		for _, s := range statuses {
			if s.Name == list[i].Name && slices.Contains(s.Installed, version) {
				list[i].Installed = true
			}
		}
	}
}

// listRollbackableNodes lists the nodes whose latest rollback point is the version.
func (h *helper) listRollbackableNodes() ([]node, error) {
	fixpack, found := cubecos.GetFixpackRawByVersion(h.reqOpts.Version)
	if !found {
		return nil, fmt.Errorf("fixpack version %s not found", h.reqOpts.Version)
	}
	if !fixpack.Rollbackable {
		return []node{}, nil
	}

	hosts, err := cubecos.ListFixpackRollbackNodes(h.reqOpts.Version)
	if err != nil {
		return nil, err
	}

	list := []nodes.Node{}
	for _, host := range hosts {
		n, err := nodes.Get(host)
		if err == nil {
			list = append(list, *n)
		}
	}

	rollbackables := h.convertToRollbackableNodes(list)
	h.sortNodesByHost(&rollbackables)
	return rollbackables, nil
}

func (h *helper) convertToUpdatableNodes(list []nodes.Node) []node {
	updatables := make([]node, 0, len(list))
	for _, n := range list {
		updatables = append(updatables, node{
			Name:      n.Hostname,
			Version:   n.Firmware.Active,
			UpdatedAt: base.ActiveFirmwareUpdatedAt,
		})
	}

	return updatables
}

func (h *helper) convertToRollbackableNodes(list []nodes.Node) []node {
	updatables := make([]node, 0, len(list))
	for _, n := range list {
		updatables = append(updatables, node{
			Name:      n.Hostname,
			UpdatedAt: base.ActiveFirmwareUpdatedAt,
		})
	}

	return updatables
}

func (h *helper) continueInterruptedFixpackUpdate() error {
	shouldReboot, err := h.checkRebootRequirement()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to check reboot requirement (%v)", h.reqId, err)
		return err
	}

	if shouldReboot {
		return h.changeNodeFixpackStatus(status.WaitingReboot)
	}

	return h.deleteReqRecord()
}

func (h *helper) deleteFixpack() error {
	err := os.Remove(h.file)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to delete fixpack file %s(%v)", h.reqId, h.file, err)
		return err
	}

	err = h.removePeerFixpacks(h.file)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to remove peer fixpacks(%v)", h.reqId, err)
		return err
	}

	return nil
}

// updateFixpackTask records one node's reported result.
func (h *helper) updateFixpackTask() error {
	switch h.reqOpts.Status.Current {
	case status.Installed, status.Rollbacked, status.Failed:
		return h.markNodeReqRecord()
	default:
		return fmt.Errorf("invalid status: %s", h.reqOpts.Status.Current)
	}
}
