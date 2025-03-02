package tunings

import (
	"errors"

	cubeHttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateTuningReq() {
	for _, role := range h.tuning.Roles {
		for _, host := range role.Hosts {
			node := host.GetNode()
			if node == nil {
				log.Errorf("failed to get node by hostname(%s)", host.Name)
				continue
			}

			if node.IsLocal() {
				delegateToLocal(h.tuning)
				continue
			}

			err := h.delegateToOtherNode(node)
			if err != nil {
				log.Errorf("failed to delegate %s to %s: %s", h.tuning.Name, node.Name, err.Error())
			}
		}
	}
}

func delegateToLocal(tuning definition.Tuning) {
	addReqRecord(tuning)
	reqQueue.Add(&tuning)
}

func (h *helper) delegateToOtherNode(node *definition.Node) error {
	url := node.PatchTuningUrl(h.tuning)
	body := h.tuning.CopyAndOverrideHost(*node)
	http := cubeHttp.GetGlobalHelper()
	resp, err := http.R().SetHeader(node.GenAuthHeader()).SetBody(body).Patch(url)
	if err != nil {
		log.Errorf("failed to send tuning %s to %s: %s", h.tuning.Name, node.Id, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to send tuning %s to %s: %d %s", h.tuning.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
