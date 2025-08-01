package images

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c         *gin.Context
	reqId     string
	handler   string
	mongo     *mongo.Helper
	openstack *openstack.Helper

	reqOpts images.ReqOpts

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

func (h *helper) listMaterials() (*materials, error) {
	projects, err := h.listProjects()
	if err != nil {
		return nil, err
	}

	dominas, err := h.listDomains()
	if err != nil {
		return nil, err
	}

	return &materials{
		ReservedImages: cubecos.GetReservedImages(),
		Projects:       projects,
		Domains:        dominas,
		Oses:           images.Oses,
		Destinations:   images.Destinations,
		Visibilities:   images.Visibilitise,
	}, nil
}

func (h *helper) listImages() ([]images.Image, error) {
	images, err := h.listConvertedImages()
	if err != nil {
		log.Errorf("images(%s): failed to list converted images(%v)", h.reqId, err)
		return nil, err
	}

	images = h.filterImages(images)
	return images, nil
}

func (h *helper) updateImageTask() error {
	switch h.reqOpts.Status.Current {
	case status.Error, status.Completed:
		return h.removePendingReq()
	default:
		return h.updatePendingReq()
	}
}
