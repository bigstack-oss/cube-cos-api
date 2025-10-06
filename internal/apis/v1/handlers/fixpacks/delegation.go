package fixpacks

import (
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/fixpacks"
)

var (
	reqQueue = fixpacks.ReqQueue
)

func (h *helper) requestOperation() {
	for _, node := range nodes.List() {
		h.addReqRecord(node.Hostname)
		if nodes.IsLocal(node.Hostname) {
			reqQueue.Add(&h.reqOpts)
		}
	}
}
