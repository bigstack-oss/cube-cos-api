package licenses

import (
	"fmt"
	"mime/multipart"
	"net/url"

	cubeHttp "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/service"
	log "go-micro.dev/v5/logger"
)

func genUrl(node definition.Node) string {
	u := url.URL{
		Scheme: "http",
		Host:   node.Address,
		Path:   fmt.Sprintf("/api/v1/datacenters/%s/nodes/%s/licenses", definition.DataCenterName, node.Role),
	}
	return u.String()
}

func sendLicenseToOtherNodes(nodeName string, licenseFile *multipart.FileHeader) error {
	// M1 TODO: might be the incorrect usage
	nodes, err := service.GetNodesByRole(nodeName)
	if err != nil {
		log.Errorf("failed to get nodes by role %s: %s", nodeName, err.Error())
		return err
	}

	reader, err := licenseFile.Open()
	if err != nil {
		log.Errorf("failed to open license file: %s", err.Error())
		return err
	}

	// M1 TODO: should protect against index out of range
	// because nodes might be empty from GetNodesByRole
	node := nodes[0]
	h := cubeHttp.GetGlobalHelper()

	resp, err := h.R().
		SetFileReader("license", licenseFile.Filename, reader).
		SetHeader("secret", "Dev@Cube").
		Post(genUrl(node))
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
