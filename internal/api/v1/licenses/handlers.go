package licenses

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

var (
	Handlers = []api.Handler{
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/licenses",
			Func:    listLicenses,
		},
		{
			Version: api.V1,
			Method:  http.MethodPost,
			Path:    "/licenses",
			Func:    importClusterLicense,
		},
		{
			Version: api.V1,
			Method:  http.MethodPost,
			Path:    "/licenses/hosts/:hostname",
			Func:    importHostLicense,
		},
	}
)

func listLicenses(c *gin.Context) {
	h, err := initHelper(c, "listLicenses")
	if err != nil {
		return
	}

	licenses, err := h.listLicenses()
	if err != nil {
		return
	}

	api.SetStatusOk(
		c,
		"fetch licenses successfully",
		licenses,
	)
}

func importClusterLicense(c *gin.Context) {
	licenseFile, err := c.FormFile("license")
	if err != nil {
		log.Errorf("request(%s): %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	filePath, err := getLicenseStorePath(licenseFile.Filename)
	if err != nil {
		log.Errorf("request(%s): failed to generate license store path: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	if err := c.SaveUploadedFile(licenseFile, filePath); err != nil {
		log.Errorf("request(%s): failed to save license file: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	if err := cubecos.ImportClusterLicense(filePath); err != nil {
		log.Errorf("request(%s): failed to import cluster license: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "update licenses successfully", nil)
}

func importHostLicense(c *gin.Context) {
	if err := importOrDelegateLicense(c, c.Param("node")); err != nil {
		log.Errorf("request(%s): failed to import license: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "update licenses successfully", nil)
}
