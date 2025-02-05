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
