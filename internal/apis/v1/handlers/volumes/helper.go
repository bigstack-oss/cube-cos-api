package volumes

import (
	"encoding/csv"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c         *gin.Context
	reqId     string
	handler   string
	mongo     *mongo.Helper
	openstack *openstack.Helper

	imageReqOpts images.ReqOpts

	project string
	page    *pages.Page
	keyword string
	watch   bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:         c,
		mongo:     mongo.GetGlobalHelper(),
		openstack: openstack.GetGlobalHelper(),
		reqId:     queries.GetReqId(c),
		handler:   handler,
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listVolumes() (*volumePage, error) {
	volumes, err := h.listConvertedVolumes()
	if err != nil {
		log.Errorf("volumes(%s): failed to list converted volumes(%v)", h.reqId, err)
		return nil, err
	}

	h.sortVolumes(&volumes)
	volumes = h.filterVolumes(volumes)
	return &volumePage{
		Volumes: h.paginateVolumes(volumes),
		Page:    h.genPageInfo(volumes),
	}, nil
}

func (h *helper) listVolumesAsCsv() (*csv.Writer, error) {
	list, err := h.listVolumes()
	if err != nil {
		return nil, err
	}

	h.c.Header("Content-Description", "File Transfer")
	h.c.Header("Content-Disposition", `attachment; filename="data.csv"`)
	h.c.Header("Content-Type", "text/csv")
	return h.convertToCsv(list.Volumes), nil
}
