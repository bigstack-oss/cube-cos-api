package fixpacks

import (
	"fmt"
	"sort"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
)

type node struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	UpdatedAt string `json:"updatedAt"`
}

type update struct {
	Version    string     `json:"version"`
	Progresses []progress `json:"progresses"`
}

type progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

func (h *helper) getUpdateDetails() (*update, error) {
	fixpack, err := cubecos.GetLastFixpackOperation()
	if err != nil {
		return nil, err
	}

	current, processPercent := h.getProgressByVersion(fixpack.Version)
	update := update{Version: fixpack.Version}
	for _, node := range nodes.List() {
		update.Progresses = append(
			update.Progresses,
			h.syncProgress(node, current, processPercent),
		)
	}

	return &update, nil
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
	if h.isVersionInstalled(version) {
		current = status.Installed
		processPercent = 100
	}

	return current, processPercent
}

func (h *helper) checkConditionForContinue() error {
	update, err := h.getFixpackUpdateProgress()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get fixpack update progress (%v)", h.reqId, err)
		return err
	}

	for _, progress := range update.Progresses {
		if progress.Status.Current == status.Failed {
			return nil
		}
	}

	return fmt.Errorf(
		"no interrupted firmware update found to continue",
	)
}
