package licenses

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	types    []string
	product  string
	products []string
	statuses []string
	roles    []string
	keyword  string
	node     string

	watch bool
	page  *pages.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, reqId: queries.GetReqId(c), handler: handler}
	return h, h.parseParamsByHandler()
}

func (h *helper) listLicenses() (*licensePage, error) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("licenses(%s): failed to list the cluster license: %s", h.reqId, err.Error())
		return nil, err
	}

	licenses = h.filterLicenses(licenses)
	return &licensePage{
		Licenses: h.paginateLicenses(licenses),
		Page:     h.genPageInfo(licenses),
	}, nil
}

func (h *helper) storeVerifyLicense() (string, error) {
	license, err := h.c.FormFile("license")
	if err != nil {
		log.Errorf("licenses(%s): %s", h.reqId, err.Error())
		return "", err
	}

	filePath, err := genLicenseVerifyPath(license.Filename)
	if err != nil {
		log.Errorf("licenses(%s): failed to generate license store path: %s", h.reqId, err.Error())
		return "", err
	}

	err = h.c.SaveUploadedFile(license, filePath)
	if err != nil {
		log.Errorf("licenses(%s): failed to save license file: %s", h.reqId, err.Error())
		return "", err
	}

	return filePath, nil
}

func (h *helper) storeImportLicense() (string, error) {
	license, err := h.c.FormFile("license")
	if err != nil {
		log.Errorf("licenses(%s): %s", h.reqId, err.Error())
		return "", err
	}

	filePath, err := genLicenseStorePath(license.Filename)
	if err != nil {
		log.Errorf("licenses(%s): failed to generate license store path: %s", h.reqId, err.Error())
		return "", err
	}

	err = h.c.SaveUploadedFile(license, filePath)
	if err != nil {
		log.Errorf("licenses(%s): failed to save license file: %s", h.reqId, err.Error())
		return "", err
	}

	return filePath, nil
}

func (h *helper) listAttachments() ([]licenses.Attachment, error) {
	attachments, err := h.listAttachmentsByProduct()
	if err != nil {
		return nil, err
	}

	return h.filterAttachments(attachments), nil
}

func (h *helper) listAttachmentsByProduct() ([]licenses.Attachment, error) {
	list, err := cubecos.ListLicenses()
	if err != nil {
		log.Errorf("licenses(%s): failed to list the licenses: %s", h.reqId, err.Error())
		return nil, err
	}

	attachments := []licenses.Attachment{}
	for _, node := range nodes.List() {
		attachment := h.genUnlicenseAttachment(node)
		status, found := h.getNodeProductStatus(node, list)
		if found {
			attachment.Status = status
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (h *helper) genUnlicenseAttachment(node nodes.Node) licenses.Attachment {
	return licenses.Attachment{
		SerialNumber: node.SerialNumber,
		Hostname:     node.Hostname,
		Role:         node.Role,
		Product:      h.normalizeProductName(h.product),
		Status:       status.Unlicense,
	}
}

func (h *helper) normalizeProductName(product string) string {
	if strings.EqualFold(product, licenses.CubeCOS) {
		return licenses.CubeCOS
	}

	if strings.EqualFold(product, licenses.CubeCMP) {
		return licenses.CubeCMP
	}

	return fmt.Sprintf(
		"Unknown Product Name(%s)",
		product,
	)
}

func (h *helper) getNodeProductStatus(node nodes.Node, licenses []licenses.License) (string, bool) {
	for _, license := range licenses {
		if !strings.EqualFold(h.product, license.Product.Name) {
			continue
		}

		if !slices.Contains(license.Hosts, node.Hostname) {
			continue
		}

		return license.Status.Current, true
	}

	return status.Unlicense, false
}
