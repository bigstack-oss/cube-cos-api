package supportfiles

import (
	"errors"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	keyword string
	host    string
	v1.SupportFile
	v1.SupportFileRequest
	v1.Page
	role string
	past string
	v1.Period

	watch bool
}

// note:
// deletion is not support in the 3.0.0 release
func initReqHandler(c *gin.Context, handler string) (*helper, error) {
	h := helper{c: c, handler: handler}
	switch h.handler {
	case "listSupportFiles":
		return initListHelper(&h)
	case "createSupportFile":
		return initCreateHelper(&h)
	case "getSupportFile":
		return initGetHelper(&h)
	case "updateSupportFileTask":
		return initUpdateHelper(&h)
	}

	return nil, errors.New("handler not found")
}

func initListHelper(h *helper) (*helper, error) {
	h.parseKeyword()
	h.parseHost()

	err := h.parsePage()
	if err != nil {
		return nil, err
	}

	err = h.parseWatch()
	if err != nil {
		return nil, err
	}

	err = h.parsePast()
	if err != nil {
		return nil, err
	}

	err = h.parsePeriod()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func initCreateHelper(h *helper) (*helper, error) {
	return h, h.parseHosts()
}

func initGetHelper(h *helper) (*helper, error) {
	return h, nil
}

func initUpdateHelper(h *helper) (*helper, error) {
	return h, nil
}

func (h *helper) listSupportFiles() (*data, error) {
	supportFiles, err := cubecos.ListSupportFiles(v1.ListSupportFileOptions{AllNodes: true})
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get supportFiles: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	supportFiles = h.filterSupportFiles(supportFiles)
	pagedSupportFiles, err := h.paginateSupportFiles(supportFiles)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to paginate supportFiles: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := h.genPageInfo(supportFiles)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to gen page info: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return &data{
		SupportFiles: pagedSupportFiles,
		Page:         page,
	}, nil
}
