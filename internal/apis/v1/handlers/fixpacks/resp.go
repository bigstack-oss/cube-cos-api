package fixpacks

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"go.mongodb.org/mongo-driver/bson"
)

type node struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	UpdatedAt string `json:"updatedAt"`
}

type progress struct {
	Host   string                      `json:"host"`
	Phase  string                      `json:"phase"`
	Status status.SystemUpdateProgress `json:"status"`
}

func (h *helper) syncProgress(node node, current string, processPercent float64) progress {
	progress := progress{
		Host: node.Name,
		Status: status.SystemUpdateProgress{
			Current:        current,
			ProcessPercent: processPercent,
		},
	}

	filter := bson.M{"hostname": node.Name, "status.current": status.Installing}
	if h.hasInprogressRecord(filter) {
		progress.Status.Current = status.Installing
		progress.Status.IsProcessing = true
		progress.Status.ProcessPercent = 50
	}

	filter["status.current"] = status.RollingBack
	if h.hasInprogressRecord(filter) {
		progress.Status.Current = status.RollingBack
		progress.Status.IsProcessing = true
		progress.Status.ProcessPercent = 50
	}

	return progress
}
