package fixpacks

import (
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	http  *http.Helper
	mongo *mongo.Helper

	file           string
	installReqOpts fixpacks.InstallReqOpts
	page           *pages.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		http:    http.GetGlobalHelper(),
		mongo:   mongo.GetGlobalHelper(),
		reqId:   queries.GetReqId(c),
		handler: handler,
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listFixpacks() (*fixpacksPage, error) {
	fixpackss, err := cubecos.ListFixpacks()
	if err != nil {
		log.Errorf("fixpackss(%s): failed to list fixpackss(%v)", h.reqId, err)
		return nil, err
	}

	h.sortFixpacks(&fixpackss)
	return &fixpacksPage{
		Fixpacks: h.paginateFixpacks(fixpackss),
		Page:     h.genPageInfo(fixpackss),
	}, nil
}
