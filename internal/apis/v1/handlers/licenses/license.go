package licenses

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/licenses"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	log "go-micro.dev/v5/logger"
)

const (
	licenseExtension  = ".license"
	licenseStorePath  = "/var/support"
	licenseVerifyPath = "/tmp/license_verify"
)

func genLicenseStorePath(filename string) (string, error) {
	filePath, err := filepath.Abs(filepath.Join(licenseStorePath, filename))
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(filePath, licenseStorePath) {
		return "", errors.New("invalid filename")
	}

	return filePath, nil
}

func genLicenseVerifyPath(filename string) (string, error) {
	filePath, err := filepath.Abs(filepath.Join(licenseVerifyPath, filename))
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(filePath, licenseVerifyPath) {
		return "", errors.New("invalid filename")
	}

	return filePath, nil
}

func (h *helper) genUnlicenseAttachmentsForAll() []licenses.Attachment {
	attachements := []licenses.Attachment{}
	for _, node := range nodes.List() {
		attachements = append(
			attachements,
			h.genUnlicenseAttachment(node),
		)
	}

	return attachements
}

func (h *helper) genAttachmentsByProduct(list []licenses.License) []licenses.Attachment {
	attachments := []licenses.Attachment{}
	for _, node := range nodes.List() {
		if node.IsNotUp() {
			tmpNode := h.getTemprorayNodeDetails(node.Hostname)
			if tmpNode != nil {
				node.SerialNumber = tmpNode.SerialNumber
				node.License = tmpNode.License
			}
		}

		attachment := h.genUnlicenseAttachment(node)
		status, found := h.getNodeProductStatus(node, list)
		if found {
			attachment.Status = status
		}

		attachments = append(attachments, attachment)
	}

	return attachments
}

func (h *helper) importOrDelegateLicense() error {
	license, err := h.c.FormFile("license")
	if err != nil {
		log.Errorf("licenses(%s): failed to get license file: %v", h.reqId, err)
		return err
	}

	if !nodes.IsLocal(h.node) {
		return h.importPeerNode(license)
	}

	return h.importLocal(license)
}

func (h *helper) importLocal(license *multipart.FileHeader) error {
	filePath, err := genLicenseStorePath(license.Filename)
	if err != nil {
		log.Errorf("licenses(%s): failed to generate license store path: %v", h.reqId, err)
		return err
	}

	err = h.c.SaveUploadedFile(license, filePath)
	if err != nil {
		log.Errorf("licenses(%s): failed to save license file: %v", h.reqId, err)
		return err
	}

	return cubecos.ImportNodeLicense(filePath)
}
