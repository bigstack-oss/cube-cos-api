package images

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	mongo   *mongo.Helper
	reqId   string
	handler string

	page  *pages.Page
	watch bool
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
