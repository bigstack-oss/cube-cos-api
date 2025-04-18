package settings

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/setting"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(setting setting.Options, err error) {
	if err != nil {
		log.Errorf("settings: failed to %s %s: %s", setting.Status.Desired, setting.Type, err.Error())
		setting.SetError()
	} else {
		log.Infof("settings: %s %s successfully", setting.Status.Desired, setting.Type)
		setting.SetCompleted()
	}

	err = o.reportToController(setting)
	if err != nil {
		return
	}
}

func (o *Operator) reportToController(setting setting.Options) error {
	node, err := v1.GetOneOfControllerNode()
	if err != nil {
		log.Errorf("settings: failed to get controller nodes: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeader(node.GenAuthHeader()).
		SetBody(setting.GenTaskUpdate()).
		Patch(node.PatchSettingTaskUrl(setting))
	if err != nil {
		log.Errorf("settings: failed to send setting %s to %s: %s", setting.Type, node.Hostname, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("settings: failed to send setting %s to %s: %v", setting.Type, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
