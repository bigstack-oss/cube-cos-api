package settings

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/settings"
	log "go-micro.dev/v5/logger"
)

func (o *Operator) handleExit(setting settings.Setting, err error) {
	if err != nil {
		log.Errorf("settings: failed to %s %s: %s", setting.Status.Desired, setting.Type, err.Error())
		setting.SetError()
	} else {
		log.Infof("settings: %s %s successfully", setting.Status.Desired, setting.Type)
		setting.SetCompleted()
	}

	if setting.IsReportRequired {
		o.reportToController(setting)
	}
}

func (o *Operator) reportToController(setting settings.Setting) {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("settings: failed to get controller nodes: %s", err.Error())
		return
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetHeaders(nodes.GetSecretHeaders()).
		SetBody(setting.GenTaskUpdate()).
		Patch(node.PatchSettingTaskUrl(setting))
	if err != nil {
		log.Errorf("settings: failed to send setting %s to %s: %s", setting.Type, node.Hostname, err.Error())
		return
	}

	if resp.IsError() {
		log.Errorf("settings: error response from %s %s update: %v", node.Hostname, setting.Type, string(resp.Body()))
	}
}
