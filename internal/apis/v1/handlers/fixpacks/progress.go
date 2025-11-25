package fixpacks

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type node struct {
	Name      string `json:"name"`
	Version   string `json:"version,omitempty"`
	UpdatedAt string `json:"updatedAt"`
}

type update struct {
	Version    string     `json:"version"`
	Operation  string     `json:"operation"`
	Progresses []progress `json:"progresses"`
}

type progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

func (h *helper) getUpdateProgressRecordByVersion(version string) (*update, error) {
	c, err := h.mongo.GetQueryCursor(
		fixpacks.Db,
		fixpacks.ReqCollection,
		bson.M{"version": version},
	)
	if err != nil {
		err := fmt.Errorf("failed to get fixpack update progress record by version %s (%v)", version, err)
		log.Errorf("fixpacks(%s): %v", h.reqId, err)
		return nil, err
	}

	if c == nil {
		err := fmt.Errorf("fixpack progress record format is unexpected")
		log.Errorf("fixpacks(%s): %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()
	defer c.Close(ctx)
	update, err := h.parseUpdateProgress(c, h.reqOpts.Version)
	if err != nil {
		return nil, err
	}

	err = h.syncRebootingDetails(update)
	if err != nil {
		return nil, err
	}

	return update, nil
}

func (h *helper) parseUpdateProgress(c *mongo.Cursor, version string) (*update, error) {
	update := update{Version: version}
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(120))
	defer cancel()
	for c.Next(ctx) {
		reqOpts := fixpacks.ReqOpts{}
		err := c.Decode(&reqOpts)
		if err != nil {
			log.Warnf("fixpacks(%s): failed to decode fixpack update progress record (%v)", h.reqId, err)
			continue
		}

		update.Version = reqOpts.Version
		update.Operation = h.convertOperationByStatus(reqOpts.Status.Current)
		current, processPercent := h.getProgressByVersion(h.reqOpts.Version)
		node, err := nodes.Get(reqOpts.Hostname)
		if err != nil {
			log.Warnf("fixpacks(%s): failed to get node %s info for fixpack progress (%v)", h.reqId, reqOpts.Hostname, err)
			continue
		}

		update.Progresses = append(
			update.Progresses,
			h.syncProgress(*node, current, processPercent),
		)
	}

	h.backfillUpdateInfo(&update)
	return &update, nil
}

func (h *helper) backfillUpdateInfo(update *update) {
	if update.Operation == "" {
		update.Operation = "waiting for update"
	}

	if update.Progresses != nil {
		return
	}

	update.Progresses = []progress{}
	for _, node := range nodes.List() {
		update.Progresses = append(
			update.Progresses,
			progress{Host: node.Hostname},
		)
	}
}

func (h *helper) syncRebootingDetails(update *update) error {
	nodes, err := cubecos.ListFixpackRebootingNodes()
	if err != nil {
		return err
	}

	rebootingMap := make(map[string]bool)
	for _, node := range nodes {
		rebootingMap[node.Hostname] = true
	}

	for i, progress := range update.Progresses {
		if progress.Status.Current == "" {
			continue
		}

		_, shouldReboot := rebootingMap[progress.Host]
		if !shouldReboot {
			continue
		}

		update.Progresses[i].Status.Current = status.WaitingReboot
		update.Progresses[i].Status.IsProcessing = true
		update.Progresses[i].Status.ProcessPercent = 100
		update.Progresses[i].Status.Description = h.getRebootingHintsByNodeRole(progress.Host)
	}

	return nil
}

func (h *helper) getRebootingHintsByNodeRole(host string) string {
	node, err := nodes.Get(host)
	if err != nil {
		return "Node needs to be rebooted to complete the fixpack update."
	}

	switch node.Role {
	case nodes.RoleControl:
		return "Control node needs to be rebooted to complete the fixpack update. Please schedule a maintenance window if necessary."
	case nodes.RoleCompute:
		return "Compute node needs to be rebooted to complete the fixpack update. Please schedule a maintenance window if necessary."
	case nodes.RoleStorage:
		return "Storage node needs to be rebooted to complete the fixpack update. Please schedule a maintenance window if necessary."
	default:
		return fmt.Sprintf("%s node needs to be rebooted to complete the fixpack update.", node.Role)
	}
}

func (h *helper) convertOperationByStatus(current string) string {
	switch strings.ToLower(current) {
	case status.Installed:
		return "install"
	case status.Rollbacked:
		return "rollback"
	default:
		log.Warnf("fixpacks(%s): unknown fixpack action %s, set operation to install by default", h.reqId, current)
		return "install"
	}
}

func (h *helper) syncProgress(node nodes.Node, current string, processPercent float64) progress {
	progress := progress{
		Host: node.Hostname,
		Status: status.SystemUpdateProgress{
			Current:        current,
			ProcessPercent: processPercent,
		},
	}

	filter := bson.M{"hostname": node.Hostname, "status.current": status.Installing}
	if h.hasInprogressUpdate(filter) {
		progress.Status.Current = status.Installing
		progress.Status.IsProcessing = true
		progress.Status.ProcessPercent = 50
	}

	filter["status.current"] = status.RollingBack
	if h.hasInprogressUpdate(filter) {
		progress.Status.Current = status.RollingBack
		progress.Status.IsProcessing = true
		progress.Status.ProcessPercent = 50
	}

	return progress
}

func (h *helper) sortUpdateProgress(progresses *[]progress) {
	sort.Slice(*progresses, func(i, j int) bool {
		return (*progresses)[i].Host < (*progresses)[j].Host
	})
}

func (h *helper) getProgressByVersion(version string) (string, float64) {
	current := status.Available
	processPercent := float64(0)
	s, err := h.getVersionStatus(version)
	if err != nil {
		return current, processPercent
	}

	switch s {
	case status.Installing:
		current = status.Installing
		processPercent = 50
	case status.Installed:
		current = status.Installed
		processPercent = 100
	case status.RollingBack:
		current = status.RollingBack
		processPercent = 50
	case status.Failed:
		current = status.Failed
		processPercent = 50
	case status.Available:
		current = status.Available
		processPercent = 0
	}

	return current, processPercent
}

func (h *helper) checkConditionForContinue() error {
	update, err := h.getFixpackUpdateProgress(h.reqOpts.Version)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get fixpack update progress (%v)", h.reqId, err)
		return err
	}

	for _, progress := range update.Progresses {
		if progress.Host != h.reqOpts.Hostname {
			continue
		}

		if progress.Status.Current == status.Failed {
			return nil
		}
	}

	return fmt.Errorf(
		"no interrupted firmware update found to continue",
	)
}
