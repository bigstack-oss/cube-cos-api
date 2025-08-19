package fixpacks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (h *helper) deleteFixpack() error {
	err := h.checkFixpackPattern()
	if err != nil {
		log.Errorf("fixpacks(%s): invalid fixpack file format (%v)", h.reqId, err)
		return err
	}

	entries, err := os.ReadDir(fixpacks.UpdateDir)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to read update directory %s(%v)", h.reqId, fixpacks.UpdateDir, err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") {
			continue
		}

		err = os.Remove(file)
		if err == nil {
			return nil
		}

		log.Errorf("fixpacks(%s): failed to delete fixpack file %s (%v)", h.reqId, file, err)
		return err
	}

	return fmt.Errorf(
		"fixpack version %s not found",
		h.file,
	)
}
