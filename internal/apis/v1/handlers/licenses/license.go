package licenses

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/gin-gonic/gin"
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

func importOrDelegateLicense(c *gin.Context, nodeName string) error {
	licenseFile, err := c.FormFile("license")
	if err != nil {
		log.Errorf("licenses(%s): failed to get license file: %s", queries.GetReqId(c), err.Error())
		return err
	}

	if nodeName != base.Hostname {
		return sendLicenseToOtherNodes(nodeName, licenseFile)
	}

	return importLicenseToNode(c, licenseFile)
}

func importLicenseToNode(c *gin.Context, licenseFile *multipart.FileHeader) error {
	filePath, err := genLicenseStorePath(licenseFile.Filename)
	if err != nil {
		log.Errorf("licenses(%s): failed to generate license store path: %s", queries.GetReqId(c), err.Error())
		return err
	}

	if err := c.SaveUploadedFile(licenseFile, filePath); err != nil {
		log.Errorf("licenses(%s): failed to save license file: %s", queries.GetReqId(c), err.Error())
		return err
	}

	return cubecos.ImportNodeLicense(filePath)
}
