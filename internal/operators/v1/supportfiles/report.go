package supportfiles

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(file *support.File, err error) {
	if err != nil {
		log.Errorf("supportfiles: failed to %s %s(%v)", file.Status.Desired, file.Group, err)
		file.SetError()
	} else {
		log.Infof("supportfiles: %s %s successfully", file.Status.Desired, file.Group)
		file.SetCompleted()
	}

	cubecos.SetSupportFileComment(*file)
	o.reportToController(*file)
}

func (o *Operator) reportToController(file support.File) {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("supportfiles: failed to get controller nodes(%v)", err)
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(file.GenTaskUpdate()).
		Patch(node.PatchSupportFileTaskUrl(file))
	if err != nil {
		log.Errorf("supportfiles: failed to send support file %s to %s(%v)", file.Group, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf(
			"supportfiles: failed to send support file %s to %s(%v)",
			file.Group,
			node.Hostname,
			string(resp.Body()),
		)
	}
}
