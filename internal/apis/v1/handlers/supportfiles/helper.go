package supportfiles

import (
	"context"
	"errors"
	"net/url"

	cubeMongo "github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type helper struct {
	c       *gin.Context
	handler string

	keyword string
	host    string
	group   support.FileSet
	file    support.File
	fileReq support.FileRequest
	pages.Page
	roles []string
	past  string
	time.Period

	watch bool
}

// note:
// deletion is not support in the 3.0.0 release
func initHepler(c *gin.Context, handler string) (*helper, error) {
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
	groupName, err := url.PathUnescape(h.c.Param("supportFileGroup"))
	if err != nil {
		return nil, err
	}

	h.group.Name = groupName
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
		log.Errorf("supportFiles(%s): failed to get supportFiles: %s", queries.GetReqId(h.c), err.Error())
		return nil, err
	}

	h.syncCreatingFile(&files)
	sets := h.convertToFileSets(files)
	pagedSets, err := h.paginateSupportFileSets(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to paginate supportFiles: %s", queries.GetReqId(h.c), err.Error())
		return nil, err
	}

	page, err := h.genPageInfo(sets)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to gen page info: %s", queries.GetReqId(h.c), err.Error())
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

func (h *helper) syncCreatingFile(files *[]support.File) {
	mongo := cubeMongo.GetGlobalHelper()
	c, err := mongo.GetQueryCursor(
		support.FileDB,
		support.FileReqCollection,
		bson.M{"status.current": "creating"},
	)
	if err != nil {
		log.Errorf("supportFiles(%s): failed to get creating file set: %s", queries.GetReqId(h.c), err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	defer c.Close(ctx)
	h.setCreatingFile(files, c)
}

func (h *helper) setCreatingFile(files *[]support.File, c *mongo.Cursor) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()
	for c.Next(ctx) {
		file := support.File{}
		err := c.Decode(&file)
		if err != nil {
			log.Errorf("supportFiles(%s): failed to decode creating file set: %s", queries.GetReqId(h.c), err.Error())
			continue
		}

		*files = append(*files, file)
	}
	if c.Err() != nil {
		log.Errorf("supportFiles(%s): failed to iterate support file cursor: %s", queries.GetReqId(h.c), c.Err().Error())
		return
	}
}
