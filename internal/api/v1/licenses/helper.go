package licenses

import (
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	Type    string
	Product string
	Status  string
	Keyword string
	Watch   bool
	v1.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{c: c, handler: handler}
	err := h.parseByHandler()
	if err != nil {
		log.Errorf("licenses(%s): failed to init request helper: %s", api.GetReqId(h.c), err.Error())
		api.SetBadRequest(c, err)
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
