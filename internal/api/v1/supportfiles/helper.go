package supportfiles

import (
	"errors"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	handler string

	keyword string
	host    string
	group   support.FileSet
	file    support.File
	fileReq support.FileRequest
	v1.Page
	roles []string
	past  string
	v1.Period

	watch bool
}

// note:
// deletion is not support in the 3.0.0 release
func initHandler(c *gin.Context, handler string) (*helper, error) {
	h := helper{c: c, handler: handler}
	switch h.handler {
	case "listSupportFiles":
		return initListHelper(&h)
	case "listHostSupportFiles":
		return initHostListHelper(&h)
	case "createSupportFile":
		return initCreateHelper(&h)
	case "downloadSupportFile":
		return initDownloadHelper(&h)
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
	h.parseRoles()

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

func initHostListHelper(h *helper) (*helper, error) {
	h.host = h.c.Param("hostname")
	return h, nil
}

func initCreateHelper(h *helper) (*helper, error) {
	return h, h.parseHosts()
}

func initDownloadHelper(h *helper) (*helper, error) {
	h.group.Name = h.c.Param("supportFileGroup")
	h.file.Name = h.c.Param("supportFileName")
	return h, nil
}

func initGetHelper(h *helper) (*helper, error) {
	return h, nil
}

func initUpdateHelper(h *helper) (*helper, error) {
	return h, h.c.ShouldBindJSON(&h.file)
}

func (h *helper) listSupportFiles() (*fileSetList, error) {
	files, err := cubecos.ListSupportFiles(support.ListFileOptions{AllNodes: true})
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get supportFiles: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	h.syncFileIsCreatedOrPending(files)
	sets := h.convertToFileSets(files)
	pagedSets, err := h.paginateSupportFileSets(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to paginate supportFiles: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := h.genPageInfo(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to gen page info: %s", api.GetReqId(h.c), err.Error())
		return nil, err
	}

	return &fileSetList{
		SupportFileSet: pagedSets,
		Page:           page,
	}, nil
}

func (h *helper) listHostSupportFiles() ([]support.File, error) {
	return cubecos.ListHostSupportFiles(support.ListFileOptions{Host: h.host})
}

func (h *helper) syncFileIsCreatedOrPending(files []support.File) {
	mongo := cubeMongo.GetGlobalHelper()
	for i, file := range files {
		count, err := mongo.GetCount(
			support.FileDB,
			support.FileReqCollection,
			genTaskFilter(file),
		)
		if err != nil {
			continue
		}

		if count > 0 {
			files[i].Status.Current = status.Creating
			files[i].Status.IsCreating = true
		}
	}
}
