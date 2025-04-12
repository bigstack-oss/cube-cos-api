package supportfiles

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateSupportFileReq() {
	for _, host := range h.fileReq.Hosts {
		node, err := v1.GetNodeByHostname(host)
		if err != nil {
			continue
		}

		h.setSupportFile()
		if node.IsLocal() {
			h.delegateToLocal()
			continue
		}

		err = h.delegateToNode(node)
		if err != nil {
			log.Errorf("supportFiles: failed to delegate %s to %s: %s", h.file.Name, node.Name, err.Error())
		}
	}
}

func (h *helper) setSupportFile() {
	if h.fileReq.CreatedAt == "" {
		h.fileReq.CreatedAt = v1.TimeISO8601Z(time.Now())
	}

	h.file = support.File{
		Group:       h.genFilSetGroup(),
		Description: h.fileReq.Description,
		Source: support.Source{
			Role: v1.CurrentRole,
			Host: v1.Hostname,
		},
		Status: status.SupportFile{
			Current:   status.Creating,
			Desired:   status.Create,
			CreatedAt: h.fileReq.CreatedAt,
		},
	}
}

func (h *helper) genFilSetGroup() string {
	return fmt.Sprintf(
		"%s Support File Set %s",
		v1.DataCenterVersion,
		h.fileReq.CreatedAt,
	)
}

func (h *helper) delegateToLocal() {
	addReqRecord(h.file)
	reqQueue.Add(&h.file)
}

func (h *helper) delegateToNode(node *v1.Node) error {
	url := node.CreateSupportFileUrl(h.file)
	body := h.genFileReqBody(*node)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeader(node.GenAuthHeader()).SetBody(body).Post(url)
	if err != nil {
		log.Errorf("failed to create support file %s to %s: %s", h.file.Name, node.Id, err.Error())
		return err
	}

	if resp.IsError() {
		log.Errorf("failed to create support file %s to %s: %d %s", h.file.Name, node.Hostname, string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func (h *helper) genFileReqBody(node v1.Node) support.FileRequest {
	return support.FileRequest{
		Name:        h.fileReq.Name,
		Description: h.fileReq.Description,
		Hosts:       []string{node.Hostname},
		CreatedAt:   h.fileReq.CreatedAt,
	}
}

func (h *helper) downloadSupportFile() error {
	setList, err := h.listSupportFiles()
	if err != nil {
		return err
	}

	if len(setList.SupportFileSet) == 0 {
		return errors.New("no support files found")
	}

	for _, set := range setList.SupportFileSet {
		for _, file := range set.Files {
			if file.Name != h.file.Name {
				continue
			}

			if file.Source.Host != v1.Hostname {
				continue
			}

			h.streamFileDownload(file.Name)
		}
	}

	return nil
}

func (h *helper) streamFileDownload(filename string) {
	file, err := os.Open(fmt.Sprintf("%s/%s", support.DefaultFileDir, filename))
	if err != nil {
		return
	}

	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return
	}

	h.c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	h.c.DataFromReader(
		nethttp.StatusOK,
		stat.Size(),
		"application/octet-stream",
		file,
		nil,
	)
}
