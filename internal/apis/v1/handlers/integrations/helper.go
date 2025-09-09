package integrations

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/integration"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listStorages() ([]integration.Storage, error) {
	cinders, err := cubecos.ListStorages()
	if err != nil {
		log.Errorf("integrations(%s): failed to list storages (%v)", h.reqId, err)
		return nil, err
	}

	storages := h.convertToStorages(cinders)
	h.sortStorages(&storages)
	return storages, nil
}
