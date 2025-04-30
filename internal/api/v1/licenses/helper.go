package licenses

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	Types    []string
	Product  string
	Products []string
	Statuses []string
	Roles    []string
	Keyword  string

	Watch bool
	v1.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	err := h.parseByHandler()
	if err != nil {
		log.Errorf("licenses(%s): failed to init request helper: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return h, nil
}

func (h *helper) listLicenses() (*data, error) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to list the cluster license: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	licenses = h.filterLicenses(licenses)
	return &data{
		Licenses: h.paginateLicenses(licenses),
		Page:     h.genPageInfo(licenses),
	}, nil
}

func (h *helper) saveLicense() (string, error) {
	licenseFile, err := h.c.FormFile("license")
	if err != nil {
		log.Errorf("license(%s): %s", api.GetReqId(h.c), err.Error())
		return "", err
	}

	filePath, err := genLicenseVerifyPath(licenseFile.Filename)
	if err != nil {
		log.Errorf("license(%s): failed to generate license store path: %s", api.GetReqId(h.c), err.Error())
		return "", err
	}

	err = h.c.SaveUploadedFile(licenseFile, filePath)
	if err != nil {
		log.Errorf("license(%s): failed to save license file: %s", api.GetReqId(h.c), err.Error())
		return "", err
	}

	return filePath, nil
}

func (h *helper) listLicenseAttachments() ([]v1.LicenseAttachment, error) {
	attachments, err := h.listLicenseAttachmentsByProduct()
	if err != nil {
		log.Warnf("licenses(%s): failed to list the license attachments: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return h.filterLicenseAttachments(attachments), nil
}

func (h *helper) listLicenseAttachmentsByProduct() ([]v1.LicenseAttachment, error) {
	licenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to list the licenses: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	nodes := v1.ListNodes()
	attachments := []v1.LicenseAttachment{}
	for _, node := range nodes {
		attachment := v1.LicenseAttachment{
			SerialNumber: node.SerialNumber,
			Hostname:     node.Hostname,
			Role:         node.Role,
			Product:      h.normalizeProductName(h.Product),
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

func (h *helper) getProductStatusOfNode(node v1.Node, licenses []v1.License) (string, bool) {
	for _, license := range licenses {
		if !strings.EqualFold(h.Product, license.Product.Name) {
			continue
		}

		if !slices.Contains(license.Hosts, node.Hostname) {
			continue
		}

		return license.Status.Current, true
	}

	return status.Unlicense, false
}
