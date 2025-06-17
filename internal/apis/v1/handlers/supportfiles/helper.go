package supportfiles

import (
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	keyword  string
	host     string
	hosts    []string
	allHosts bool
	group    support.FileSet
	file     support.File
	fileReq  support.FileRequest
	roles    []string

	*pages.Page
	past string
	*time.Period

	watch bool
}

func initHepler(c *gin.Context, handler string) (*helper, error) {
	h := helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
		Page:    &pages.Page{},
	}

	return &h, h.parseParamsByHandler()
}

func (h *helper) listSupportFiles() (*filePage, error) {
	sets, err := h.listSupportFileSets()
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list file sets: %v", h.reqId, err)
		return nil, err
	}

	pagedSets, err := h.paginateFileSets(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to paginate files: %v", h.reqId, err)
		return nil, err
	}

	page, err := h.genPageInfo(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to gen page info: %v", h.reqId, err)
		return nil, err
	}

	return &filePage{
		SupportFileSet: pagedSets,
		Page:           page,
	}, nil
}

func (h *helper) listSupportFileSets() ([]support.FileSet, error) {
	files, err := cubecos.ListSupportFiles(support.ListFileOptions{AllNodes: true})
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list files: %v", h.reqId, err)
		return nil, err
	}

	h.syncCreatingFiles(&files)
	h.syncHostPortInUrl(&files)
	return h.genFileSets(files), nil
}
