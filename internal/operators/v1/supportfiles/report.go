package supportfiles

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(file *support.File, err error) {
	if err != nil {
		log.Errorf("supportfiles: failed to %s %s: %s", file.Status.Desired, file.Group, err.Error())
		file.SetError()
	} else {
		log.Infof("supportfiles: %s %s successfully", file.Status.Desired, file.Group)
		file.SetCompleted()
	}

	cubecos.SetSupportFileComment(*file)
	err = o.reportToController(*file)
	if err != nil {
		return
	}
}

func (o *Operator) reportToController(file support.File) error {
	node, err := definition.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("supportfiles: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeader(node.GenAuthHeader()).
		SetBody(file.GenTaskUpdate()).
		Patch(node.PatchSupportFileTaskUrl(file))
	if err != nil {
		log.Errorf("supportfiles: failed to send support file %s to %s: %s", file.Group, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("supportfiles: failed to send support file %s to %s: %v", file.Group, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
