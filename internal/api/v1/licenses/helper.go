package licenses

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/license"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	types    []string
	product  string
	products []string
	statuses []string
	roles    []string
	keyword  string

	watch bool
	page  *v1.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	return h, h.parseParamsByHandler()
}

func (h *helper) listLicenses() (*licensePages, error) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("licenses(%s): failed to list the cluster license: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	licenses = h.filterLicenses(licenses)
	return &licensePages{
		Licenses: h.paginateLicenses(licenses),
		Page:     h.genPageInfo(licenses),
	}, nil
}

func (h *helper) saveLicense() (string, error) {
	license, err := h.c.FormFile("license")
	if err != nil {
		log.Errorf("licenses(%s): %s", api.GetReqId(h.c), err.Error())
		return "", err
	}

	filePath, err := genLicenseVerifyPath(license.Filename)
	if err != nil {
		log.Errorf("licenses(%s): failed to generate license store path: %s", api.GetReqId(h.c), err.Error())
		return "", err
	}

	err = h.c.SaveUploadedFile(license, filePath)
	if err != nil {
		log.Errorf("licenses(%s): failed to save license file: %s", api.GetReqId(h.c), err.Error())
		return "", err
	}

	return filePath, nil
}

func (h *helper) listAttachments() ([]license.Attachment, error) {
	attachments, err := h.listAttachmentsByProduct()
	if err != nil {
		log.Warnf("licenses(%s): failed to list the license attachments: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return h.filterAttachments(attachments), nil
}

func (h *helper) listAttachmentsByProduct() ([]license.Attachment, error) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("licenses(%s): failed to list the licenses: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	attachments := []license.Attachment{}
	for _, node := range nodes.List() {
		attachment := license.Attachment{
			SerialNumber: node.SerialNumber,
			Hostname:     node.Hostname,
			Role:         node.Role,
			Product:      h.normalizeProductName(h.product),
			Status:       status.Unlicense,
		}

		status, found := h.getProductStatusOfNode(node, licenses)
		if found {
			attachment.Status = status
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (h *helper) normalizeProductName(product string) string {
	if strings.EqualFold(product, "CubeCOS") {
		return "CubeCOS"
	}

	if strings.EqualFold(product, "CubeCMP") {
		return "CubeCMP"
	}

	return fmt.Sprintf(
		"Unknown Product Name(%s)",
		product,
	)
}

func (h *helper) getProductStatusOfNode(node nodes.Node, licenses []license.Options) (string, bool) {
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
