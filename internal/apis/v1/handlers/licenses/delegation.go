package licenses

import (
	"fmt"
	"mime/multipart"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func (h *helper) importPeerNode(license *multipart.FileHeader) error {
	node, err := nodes.Get(h.node)
	if err != nil {
		log.Errorf("license(%s): failed to get nodes by role %s: %s", queries.GetReqId(h.c), h.node, err.Error())
		return err
	}

	reader, err := license.Open()
	if err != nil {
		log.Errorf("licenses(%s): failed to open license file: %s", queries.GetReqId(h.c), err.Error())
		return err
	}

	http := http.GetGlobalHelper()
	resp, err := http.R().
		SetFileReader("license", license.Filename, reader).
		SetHeaders(nodes.GetSecretHeaders()).
		Post(node.PostLicenseUrl())
	if err != nil {
		return err
	}

	if !resp.IsError() {
		return nil
	}

	err = fmt.Errorf("has resp error for license: %s", string(resp.Body()))
	log.Errorf("licenses(%s): %s", queries.GetReqId(h.c), err.Error())
	return err
}
