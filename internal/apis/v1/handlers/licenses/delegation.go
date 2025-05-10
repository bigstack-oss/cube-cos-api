package licenses

import (
	"fmt"
	"mime/multipart"
	"net/url"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

func genUrl(node nodes.Node) string {
	u := url.URL{
		Scheme: "http",
		Host:   node.Address,
		Path:   fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/licenses", base.DataCenterName, node.Role),
	}
	return u.String()
}

func sendLicenseToOtherNodes(nodeName string, licenseFile *multipart.FileHeader) error {
	node, err := nodes.GetController()
	if err != nil {
		log.Errorf("failed to get nodes by role %s: %s", nodeName, err.Error())
		return err
	}

	reader, err := licenseFile.Open()
	if err != nil {
		log.Errorf("failed to open license file: %s", err.Error())
		return err
	}

	h := http.GetGlobalHelper()
	resp, err := h.R().
		SetFileReader("license", licenseFile.Filename, reader).
		SetHeaders(nodes.GetSecretHeaders()).
		Post(genUrl(*node))
	if resp.IsError() || err != nil {
		log.Errorf(
			"failed to send license to node %s: %d %s",
			node.Hostname,
			resp.StatusCode(),
			string(resp.Body()),
		)
		return err
	}

	return nil
}
