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
			Func:    getLicenses,
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
			Path:    "/nodes/:node/licenses",
			Func:    importNodeLicense,
		},
	}
)

func getLicenses(c *gin.Context) {
	pageOpts, err := genPageOptsByQueryParams(c)
	if err != nil {
		log.Errorf("request(%s): %s", api.GetReqId(c), err.Error())
		api.SetBadRequest(c, err)
		return
	}

	allLicenses, err := cubecos.ListLicenses()
	if err != nil {
		log.Warnf("request(%s): failed to list the cluster license: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	pagedLicenses, err := paginateLicenses(allLicenses, pageOpts)
	if err != nil {
		log.Errorf("request(%s): failed to paginate licenses: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	page, err := genPageInfo(pagedLicenses, pageOpts)
	if err != nil {
		log.Errorf("request(%s): failed to gen page info: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(
		c,
		"fetch licenses successfully",
		data{
			Licenses: pagedLicenses,
			Page:     page,
		},
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

func importNodeLicense(c *gin.Context) {
	if err := importOrDelegateLicense(c, c.Param("node")); err != nil {
		log.Errorf("request(%s): failed to import license: %s", api.GetReqId(c), err.Error())
		api.SetInternalServerError(c, err)
		return
	}

	api.SetStatusOk(c, "update licenses successfully", nil)
}
