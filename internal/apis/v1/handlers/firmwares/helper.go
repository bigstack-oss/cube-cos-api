package firmwares

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string
	mongo   *mongo.Helper

	file string
	page *pages.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		mongo:   mongo.GetGlobalHelper(),
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listFirmwares() (*firmwarePage, error) {
	firmwares, err := cubecos.ListFirmwares()
	if err != nil {
		log.Errorf("images(%s): failed to list converted images(%v)", h.reqId, err)
		return nil, err
	}

	h.sortFirmwares(&firmwares)
	return &firmwarePage{
		Firmwares: h.paginateFirmwares(firmwares),
		Page:      h.genPageInfo(firmwares),
	}, nil
}
