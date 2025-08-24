package fixpacks

import (
	"fmt"
	"os"
	"sort"

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
	"go.mongodb.org/mongo-driver/bson"
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

func (h *helper) listFixpacks() (*fixpacksPage, error) {
	fixpackss, err := cubecos.ListFixpacks()
	if err != nil {
		log.Errorf("fixpackss(%s): failed to list fixpackss(%v)", h.reqId, err)
		return nil, err
	}

	h.sortFixpacks(&fixpackss)
	h.syncRequestingRecord(&fixpackss)
	return &fixpacksPage{
		Fixpacks: h.paginateFixpacks(fixpackss),
		Page:     h.genPageInfo(fixpackss),
	}, nil
}

func (h *helper) listFixpackUpdateProgress(version string) ([]progress, error) {
	updatables, err := h.listUpdatableNodes(version)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to list updatable nodes for fixpack %s(%v)", h.reqId, version, err)
		return nil, err
	}

	current := status.Available
	processPercent := float64(0)
	if h.isVersionInstalled(version) {
		current = status.Installed
		processPercent = 100
	}

	progresses := []progress{}
	for _, node := range updatables {
		progress := progress{
			Host: node.Name,
			Status: status.FixpackProgress{
				Current:        current,
				ProcessPercent: processPercent,
			},
		}

		filter := bson.M{"hostname": node.Name, "status.current": status.Installing}
		if h.hasInprogressRecord(filter) {
			progress.Status.Current = status.Installing
			progress.Status.IsProcessing = true
			progress.Status.ProcessPercent = 50
			progresses = append(progresses, progress)
			continue
		}

		filter["status.current"] = status.RollingBack
		if h.hasInprogressRecord(filter) {
			progress.Status.Current = status.RollingBack
			progress.Status.IsProcessing = true
			progress.Status.ProcessPercent = 50
			progresses = append(progresses, progress)
			continue
		}

		progresses = append(progresses, progress)
	}

	sort.Slice(progresses, func(i, j int) bool {
		return (progresses)[i].Host < (progresses)[j].Host
	})

	return progresses, nil
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

	h.sortUpdatableNodes(&updatables)
	return updatables, nil
}

func (h *helper) installFixpack(updatables []node) error {
	err := h.syncFixpack(updatables)
	if err != nil {
		return err
	}

	h.delegateToLocal(updatables)
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return nil
	}

	h.delegateToPeers(updatables)
	return nil
}

func (h *helper) rollbackFixpack(updateds []node) error {
	h.delegateToLocal(updateds)
	if !cubecos.IsVirtualIpOwner(base.Hostname) {
		return nil
	}

	h.delegateToPeers(updateds)
	return nil
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

func (h *helper) continueInterruptedFixpackUpdate() error {
	node, err := cubecos.GetUpdateInterruptedNode()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get interrupted nodes (%v)", h.reqId, err)
		return err
	}

	if node.IsVirtualIpOwner {
		cubecos.MoveVirtualIpOwner()
	}

	// have to check if the fixpack is not required to be rebooted,
	// then just delete the req record and return
	// if !fixpack.IsRebootRequired() {
	// 	h.deleteReqRecord()
	// 	return nil
	// }

	err = cubecos.SoftRebootBySsh(node.Hostname)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to soft reboot node %s (%v)", h.reqId, node.Hostname, err)
		return err
	}

	return nil
}

func (h *helper) deleteFixpack() error {
	err := os.Remove(h.file)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to delete fixpack file %s(%v)", h.reqId, h.file, err)
		return err
	}

	return nil
}

func (h *helper) updateFixpackTask() error {
	switch h.reqOpts.Status.Current {
	case status.Completed:
		return h.deleteReqRecord()
	case status.Error:
		return h.markReqRecordAsFailed()
	default:
		return fmt.Errorf("invalid status: %s", h.reqOpts.Status.Current)
	}
}
