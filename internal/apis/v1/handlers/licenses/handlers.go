package licenses

import (
	"net/http"

	api "github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
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
			Func:    listAttachments,
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

	bodies.SetOk(
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

	license, err := h.storeVerifyLicense()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	result, err := cubecos.VerifyLicense(license)
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	bodies.SetOk(
		c,
		"license verified successfully",
		result,
	)
}

func importClusterLicense(c *gin.Context) {
	h, err := initHelper(c, "importClusterLicense")
	if err != nil {
		log.Errorf("license(%s): failed to init helper: %s", queries.GetReqId(c), err.Error())
		bodies.SetBadRequest(c, err)
		return
	}

	path, err := h.storeImportLicense()
	if err != nil {
		bodies.SetBadRequest(c, err)
		return
	}

	err = cubecos.ImportClusterLicense(path)
	if err != nil {
		log.Errorf("license(%s): failed to import cluster license: %s", queries.GetReqId(c), err.Error())
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"import license successfully",
		nil,
	)
}

func importHostLicense(c *gin.Context) {
	h, err := initHelper(c, "importHostLicense")
	if err != nil {
		log.Errorf("license(%s): failed to init helper: %s", queries.GetReqId(c), err.Error())
		bodies.SetBadRequest(c, err)
		return
	}

	err = h.importOrDelegateLicense()
	if err != nil {
		log.Errorf("license(%s): failed to import license: %s", queries.GetReqId(c), err.Error())
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"update licenses successfully",
		nil,
	)
}

func listAttachments(c *gin.Context) {
	h, err := initHelper(c, "listAttachments")
	if err != nil {
		log.Errorf("license(%s): failed to init helper: %s", queries.GetReqId(c), err.Error())
		bodies.SetBadRequest(c, err)
		return
	}

	attachments, err := h.listAttachments()
	if err != nil {
		bodies.SetInternalServerError(c, err)
		return
	}

	bodies.SetOk(
		c,
		"fetch license attachments successfully",
		attachments,
	)
}
