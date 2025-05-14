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

	keyword string
	host    string
	group   support.FileSet
	file    support.File
	fileReq support.FileRequest
	roles   []string

	*pages.Page
	past string
	*time.Period

	watch bool
}

// note:
// deletion will be supported in the M2 release
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
	files, err := cubecos.ListSupportFiles(support.ListFileOptions{AllNodes: true})
	if err != nil {
		log.Errorf("supportFiles(%s): failed to list files: %v", h.reqId, err)
		return nil, err
	}

	h.syncCreatingFiles(&files)
	h.syncHostPortInUrl(&files)
	sets := h.genFileSets(files)
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
