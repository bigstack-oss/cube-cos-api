package licenses

import (
	"net/http"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	_ "github.com/bigstack-oss/cube-cos-api/internal/operators/v1/licenses"
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
			Path:    "/licenses/verify",
			Func:    verifyLicense,
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
		{
			Version: api.V1,
			Method:  http.MethodGet,
			Path:    "/licenses/attachments",
			Func:    listLicenseAttachments,
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

func verifyLicense(c *gin.Context) {
	h, err := initHelper(c, "verifyLicense")
	if err != nil {
		return
	}

	license, err := h.saveLicense()
	if err != nil {
		api.SetBadRequest(c, err)
		return
	}

	result, err := cubecos.VerifyLicense(license)
	if err != nil {
		api.SetBadRequest(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"license verified successfully",
		result,
	)
}

func importClusterLicense(c *gin.Context) {
	licenseFile, err := c.FormFile("license")
	if err != nil {
		log.Errorf("license(%s): %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	filePath, err := genLicenseStorePath(licenseFile.Filename)
	if err != nil {
		log.Errorf("license(%s): failed to generate license store path: %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	err = c.SaveUploadedFile(licenseFile, filePath)
	if err != nil {
		log.Errorf("license(%s): failed to save license file: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	err = cubecos.ImportClusterLicense(filePath)
	if err != nil {
		log.Errorf("license(%s): failed to import cluster license: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "update licenses successfully", nil)
}

func importHostLicense(c *gin.Context) {
	if err := importOrDelegateLicense(c, c.Param("node")); err != nil {
		log.Errorf("license(%s): failed to import license: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "update licenses successfully", nil)
}

func listLicenseAttachments(c *gin.Context) {
	h, err := initHelper(c, "listLicenseAttachments")
	if err != nil {
		return
	}

	attachments, err := h.listLicenseAttachments()
	if err != nil {
		return
	}

	api.SetStatusOk(
		c,
		"fetch license attachments successfully",
		attachments,
	)
}
