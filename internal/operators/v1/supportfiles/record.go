package supportfiles

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(supportFile v1.SupportFile, err error) {
	if err != nil {
		log.Errorf("supportfiles: failed to %s %s: %s", supportFile.Status.Desired, supportFile.Name, err.Error())
		supportFile.SetError()
	} else {
		log.Infof("supportfiles: %s %s successfully", supportFile.Status.Desired, supportFile.Name)
		supportFile.SetCompleted()
	}

	err = o.reportToController(supportFile)
	if err != nil {
		return
	}
}

func (o *Operator) reportToController(supportFile v1.SupportFile) error {
	node, err := definition.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("supportfiles: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeader(node.GenAuthHeader()).
		SetBody(supportFile.GenTaskUpdate()).
		Patch(node.PatchSupportFileTaskUrl(supportFile))
	if err != nil {
		log.Errorf("supportfiles: failed to send support file %s to %s: %s", supportFile.Name, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("supportfiles: failed to send support file %s to %s: %v", supportFile.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
