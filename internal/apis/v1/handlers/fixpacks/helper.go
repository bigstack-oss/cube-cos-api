package fixpacks

import (
	"fmt"
	"os"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
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

func (h *helper) listFixpacks() (*fixpacksPage, error) {
	fixpackss, err := cubecos.ListFixpacks()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to list fixpackss(%v)", h.reqId, err)
		return nil, err
	}

	h.syncRequestingRecord(&fixpackss)
	return &fixpacksPage{
		Fixpacks: h.paginateFixpacks(fixpackss),
		Page:     h.genPageInfo(fixpackss),
	}, nil
}

func (h *helper) getFixpackUpdateProgress() (*update, error) {
	update, err := h.getUpdateDetails()
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

	h.sortNodesByHost(&updatables)
	return updatables, nil
}

func (h *helper) listRollbackableNodes() ([]node, error) {
	fixpack, found := cubecos.GetFixpackRawByVersion(h.reqOpts.Version)
	if !found {
		return nil, fmt.Errorf("fixpack version %s not found", h.reqOpts.Version)
	}
	if fixpack.NoRollback {
		return []node{}, nil
	}

	list, err := h.filterNodesByRole(fixpack.RebootRequired)
	if err != nil {
		return nil, err
	}

	updatables := h.convertToRollbackableNodes(list)
	h.sortNodesByHost(&updatables)
	return updatables, nil
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
	node, err := nodes.Get(h.reqOpts.Hostname)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get node %s (%v)", h.reqId, h.reqOpts.Hostname, err)
		return err
	}

	if node.IsVirtualIpOwner {
		cubecos.MoveVirtualIpOwner()
	}

	shouldReboot, err := h.checkRebootRequirement()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to check reboot requirement (%v)", h.reqId, err)
		return err
	}

	h.deleteReqRecord()
	if !shouldReboot {
		return nil
	}

	return cubecos.SoftRebootBySsh(node.Hostname)
}

func (h *helper) deleteFixpack() error {
	err := os.Remove(h.file)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to delete fixpack file %s(%v)", h.reqId, h.file, err)
		return err
	}

	return nil
}

func (h *helper) updateFixpackTask(nodes []node) error {
	switch h.reqOpts.Status.Current {
	case status.Completed:
		return h.deleteReqRecord()
	case status.Failed:
		failures := h.findFailedNodes(nodes)
		return h.markReqRecordAsFailed(failures)
	default:
		return fmt.Errorf("invalid status: %s", h.reqOpts.Status.Current)
	}
}

func (h *helper) findFailedNodes(list []node) []nodes.Node {
	failures := []nodes.Node{}

	for _, n := range list {
		node, err := nodes.Get(n.Name)
		if err != nil {
			log.Warnf("fixpacks(%s): failed to get node %s (%v)", h.reqId, n.Name, err)
			continue
		}

		fixpack := &fixpacks.Fixpack{}
		if node.IsLocal() {
			fixpack, err = cubecos.GetLatestFixpackInfo()
		} else {
			fixpack, err = h.askPeerFixpackInfo(*node)
		}
		if err != nil {
			continue
		}

		if fixpack.Version != h.reqOpts.Version {
			failures = append(failures, *node)
		}
	}

	return failures
}

func (h *helper) askPeerFixpackInfo(node nodes.Node) (*fixpacks.Fixpack, error) {
	resp, err := h.http.R().
		SetResult(&bodies.Fixpack{}).
		SetHeaders(nodes.GetSecretHeaders()).
		Get(node.GetFixpackInfoUrl())
	if err != nil {
		log.Errorf("fixpacks: failed to get node details %s: %v", node.Hostname, err)
		return nil, err
	}

	if !resp.IsError() {
		return &resp.Result().(*bodies.Fixpack).Data, nil
	}

	err = fmt.Errorf("resp error for node fixpack info %s: %s", node.Hostname, string(resp.Body()))
	log.Errorf("fixpacks(%v)", err)
	return nil, err
}
