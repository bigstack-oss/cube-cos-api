package supportfiles

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

var ()

func (h *helper) delegateSupportFileReq() {
	h.supportFile.Comment = v1.TimeLocal()
	for _, role := range h.supportFile.Roles {
		for _, node := range role.Nodes {
			h.supportFile.InitCreateStatus()
			if node.IsLocal() {
				delegateToLocal(h.supportFile.GenTask(*node))
				continue
			}

			err := h.delegateToOtherNode(node)
			if err != nil {
				log.Errorf("supportFiles: failed to delegate %s to %s: %s", h.supportFile.Name, node.Name, err.Error())
			}
		}
	}
}

func delegateToLocal(supportFile v1.SupportFile) {
	addReqRecord(supportFile)
	reqQueue.Add(&supportFile)
}

func (h *helper) delegateToOtherNode(node *v1.Node) error {
	url := node.CreateSupportFileUrl(h.supportFile)
	body := h.supportFile.GenTask(*node)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeader(node.GenAuthHeader()).SetBody(body).Post(url)
	if err != nil {
		log.Errorf("failed to create support file %s to %s: %s", h.supportFile.Name, node.Id, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to create support file %s to %s: %d %s", h.supportFile.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
