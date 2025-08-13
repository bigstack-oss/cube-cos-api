package supportfiles

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	ostime "time"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	log "go-micro.dev/v5/logger"
)

func (h *helper) delegateSupportFileReq() {
	for _, host := range h.fileReq.Hosts {
		node, err := nodes.Get(host)
		if err != nil {
			continue
		}

		if node.IsLocal() {
			h.updateToLocal()
			continue
		}

		if node.IsDown() {
			log.Errorf(
				"supportFiles(%s): node %s is down, cannot delegate %s",
				h.reqId,
				node.Hostname,
				h.file.Name,
			)
			continue
		}

		err = h.tunePeer(node)
		if err != nil {
			log.Errorf(
				"supportFiles(%s): failed to delegate %s to %s: %v",
				h.reqId,
				h.file.Name,
				node.Hostname,
				err,
			)
		}
	}
}

func (h *helper) setSupportFileReq() {
	if h.fileReq.CreatedAt == "" {
		h.fileReq.CreatedAt = time.ISO8601Z(ostime.Now())
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
		base.DataCenterFirmwareVersion,
		h.fileReq.CreatedAt,
	)
}

func (h *helper) updateToLocal() {
	h.addReqRecord(h.file)
	reqQueue.Add(&h.file)
}

func (h *helper) tunePeer(node *nodes.Node) error {
	url := node.CreateSupportFileUrl(h.file)
	body := h.genFileReqBody(*node)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeaders(nodes.GetSecretHeaders()).SetBody(body).Post(url)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to create support file %s to %s: %v", h.reqId, h.file.Name, node.Id, err)
		return err
	}

	if resp.IsError() {
		log.Errorf("supportFiles(%s): failed to create support file %s to %s: %s", h.reqId, h.file.Name, node.Hostname, string(resp.Body()))
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
	list, err := h.listSupportFiles()
	if err != nil {
		return err
	}
	if len(list.SupportFileSet) == 0 {
		return errors.New("no support files found")
	}

	set := h.findFileSet(list.SupportFileSet)
	for _, file := range set.Files {
		if file.Name != h.file.Name {
			continue
		}

		if file.Source.Host == base.Hostname {
			h.streamFileDownload(file.Name)
			break
		}

		h.streamDownloadByPeerNode(set, file)
		break
	}

	return nil
}

func (h *helper) findFileSet(sets []support.FileSet) support.FileSet {
	for _, set := range sets {
		if set.Name == h.group.Name {
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

func (h *helper) deleteSupportFileGroup() error {
	hosts, err := h.getHostsByGroup(h.group.Name)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get hosts by group %s: %v", h.reqId, h.group.Name, err)
		return err
	}

	return h.deleteSupportFiles(hosts)
}

func (h *helper) deleteSupportFiles(hosts []string) error {
	for _, host := range hosts {
		file, err := h.GetHostSupportFile(h.group.Name, host)
		if err != nil {
			log.Warnf("supportFiles(%s): failed to get support file by group %s(%v)", h.reqId, h.group.Name, err)
			continue
		}

		if file.IsCreating() {
			log.Warnf("supportFiles(%s): group %s has creating file, skip to delete", h.reqId, h.group.Name)
			return errors.New("has creating file, skip to delete")
		}

		node, err := nodes.Get(host)
		if err != nil {
			log.Errorf("supportFiles(%s): failed to get node by hostname %s(%v)", h.reqId, host, err)
			continue
		}

		if node.IsLocal() {
			cubecos.DeleteSupportFile(*file)
			continue
		}

		if node.IsDown() {
			log.Warnf("supportFiles(%s): node %s is down, cannot delete group %s", h.reqId, node.Hostname, h.group.Name)
			continue
		}

		h.deletePeerFile(node, file.Name)
	}

	return nil
}

func (h *helper) deletePeerFile(node *nodes.Node, filename string) {
	url := node.DeleteSupportFileUrl(h.group.Name, filename)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeaders(nodes.GetSecretHeaders()).Delete(url)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to delete support file group %s from %s(%v)", h.reqId, h.group.Name, node.Hostname, err)
		return
	}

	if resp.IsError() {
		log.Errorf("supportFiles(%s): %s resp error from %s(%s)", h.reqId, h.group.Name, node.Hostname, string(resp.Body()))
		return
	}

	log.Infof(
		"supportFiles(%s): support file group(%s) deleted successfully from %s",
		h.reqId,
		h.group.Name,
		node.Hostname,
	)
}

func (h *helper) GetHostSupportFile(group, host string) (*support.File, error) {
	sets, err := h.listSupportFileSets()
	if err != nil {
		return nil, err
	}

	for _, set := range sets {
		if set.Name != group {
			continue
		}

		file, found := h.findFile(set, host)
		if found {
			return file, nil
		}
	}

	return nil, fmt.Errorf(
		"support file group(%s) not found",
		group,
	)
}

func (h *helper) findFile(set support.FileSet, host string) (*support.File, bool) {
	for _, file := range set.Files {
		if file.Source.Host != host {
			continue
		}

		return &file, true
	}

	return nil, false
}

func (h *helper) streamDownloadByPeerNode(set support.FileSet, file support.File) {
	node, err := nodes.Get(file.Source.Host)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get node by hostname %s: %v", h.reqId, file.Source.Host, err)
		return
	}

	url := node.DownloadSupportFileUrl(set.Name, file.Name)
	http := http.GetGlobalHelper()
	resp, err := http.R().SetHeaders(nodes.GetSecretHeaders()).Get(url)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to download support file %s from %s: %v", h.reqId, file.Name, node.Hostname, err)
		return
	}
	if resp.IsError() {
		log.Errorf("supportFiles(%s): %s resp error from %s: %s", h.reqId, file.Name, node.Hostname, string(resp.Body()))
		return
	}

	h.c.Data(
		nethttp.StatusOK,
		"application/octet-stream",
		resp.Body(),
	)
}
