package supportfiles

import (
	"errors"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/operators/v1/supportfiles"
	log "go-micro.dev/v5/logger"
)

var (
	reqQueue = supportfiles.ReqQueue
)

func (h *helper) delegateSupportFileReq() {
	h.SupportFile.Group = v1.TimeLocal()
	for _, role := range h.SupportFile.Roles {
		for _, node := range role.Nodes {
			h.SupportFile.InitCreateStatus()
			if node.IsLocal() {
				delegateToLocal(h.SupportFile.GenTask(*node))
				continue
			}

			err := h.delegateToOtherNode(node)
			if err != nil {
				log.Errorf("failed to delegate %s to %s: %s", h.SupportFile.Name, node.Name, err.Error())
			}
		}
	}
}

func delegateToLocal(supportFile v1.SupportFile) {
	addReqRecord(supportFile)
	reqQueue.Add(&supportFile)
}

func (h *helper) delegateToOtherNode(node *v1.Node) error {
	url := node.CreateSupportFileUrl(h.SupportFile)
	body := h.SupportFile.GenTask(*node)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeader(node.GenAuthHeader()).SetBody(body).Post(url)
	if err != nil {
		log.Errorf("failed to create support file %s to %s: %s", h.SupportFile.Name, node.Id, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to create support file %s to %s: %d %s", h.SupportFile.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
