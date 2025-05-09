package supportfiles

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auth"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateSupportFileReq() {
	for _, host := range h.fileReq.Hosts {
		node, err := nodes.Get(host)
		if err != nil {
			continue
		}

		h.setSupportFile()
		if node.IsLocal() {
			h.delegateToLocal()
			continue
		}

		if node.IsDown() {
			log.Errorf("supportFiles: node %s is down, cannot delegate %s", node.Hostname, h.file.Name)
			continue
		}

		err = h.delegateToNode(node)
		if err != nil {
			log.Errorf("supportFiles: failed to delegate %s to %s: %s", h.file.Name, node.Hostname, err.Error())
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
			Role: base.CurrentRole,
			Host: base.Hostname,
		},
		Status: status.SupportFile{
			Current:    status.Creating,
			Desired:    status.Create,
			CreatedAt:  h.fileReq.CreatedAt,
			IsCreating: true,
		},
	}
}

func (h *helper) genFilSetGroup() string {
	return fmt.Sprintf(
		"%s Support File Set %s",
		base.DataCenterVersion,
		h.fileReq.CreatedAt,
	)
}

func (h *helper) delegateToLocal() {
	h.addReqRecord(h.file)
	reqQueue.Add(&h.file)
}

func (h *helper) delegateToNode(node *nodes.Node) error {
	url := node.CreateSupportFileUrl(h.file)
	body := h.genFileReqBody(*node)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeaders(auth.GetNodeSecret()).SetBody(body).Post(url)
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

func (h *helper) genFileReqBody(node nodes.Node) support.FileRequest {
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

	set := h.findFileSet(setList.SupportFileSet)
	for _, file := range set.Files {
		if file.Name != h.file.Name {
			continue
		}

		if file.Source.Host == base.Hostname {
			h.streamFileDownload(file.Name)
			break
		}

		h.streamFromPeerNode(set, file)
		break
	}

	return nil
}

func (h *helper) findFileSet(sets []support.FileSet) support.FileSet {
	for _, set := range sets {
		if set.Name == h.group.Name {
			log.Infof("found file set: %s", set.Name)
			return set
		}
	}

	return support.FileSet{}
}

func (h *helper) streamFileDownload(filename string) {
	filepath := fmt.Sprintf("%s/%s", support.DefaultFileDir, filename)
	file, err := os.Open(filepath)
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

func (h *helper) streamFromPeerNode(set support.FileSet, file support.File) {
	node, err := nodes.Get(file.Source.Host)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get node by hostname %s: %s", api.GetReqId(h.c), file.Source.Host, err.Error())
		return
	}

	url := node.DownloadSupportFileUrl(set.Name, file.Name)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeaders(auth.GetNodeSecret()).Get(url)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to download support file %s from %s: %s", api.GetReqId(h.c), file.Name, node.Hostname, err.Error())
		return
	}
	if resp.IsError() {
		log.Errorf("supportFiles(%s): %s resp error from %s: %d %s", api.GetReqId(h.c), file.Name, node.Hostname, resp.StatusCode(), string(resp.Body()))
		return
	}

	h.c.Data(
		nethttp.StatusOK,
		"application/octet-stream",
		resp.Body(),
	)
}
