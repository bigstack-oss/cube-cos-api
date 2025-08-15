package firmwares

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string
	mongo   *mongo.Helper

	file string
	page *pages.Page
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

func (h *helper) listFirmwares() (*firmwarePage, error) {
	firmwares, err := cubecos.ListFirmwares()
	if err != nil {
		log.Errorf("firmwares(%s): failed to list firmwares(%v)", h.reqId, err)
		return nil, err
	}

	h.sortFirmwares(&firmwares)
	return &firmwarePage{
		Firmwares: h.paginateFirmwares(firmwares),
		Page:      h.genPageInfo(firmwares),
	}, nil
}

func (h *helper) deleteFirmware() error {
	err := h.checkFirmwarePattern()
	if err != nil {
		return err
	}

	segments := strings.Split(h.file, " ")
	version := segments[2]
	hash := segments[3]
	entries, err := os.ReadDir(firmwares.UpdateDir)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read update directory %s(%v)", h.reqId, firmwares.UpdateDir, err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(firmwares.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".pkg") {
			continue
		}

		if !strings.Contains(file, version) {
			continue
		}

		if !strings.Contains(file, hash) {
			continue
		}

		err = os.Remove(file)
		if err == nil {
			return nil
		}

		log.Errorf("firmwares(%s): failed to delete firmware file %s (%v)", h.reqId, file, err)
		return err
	}

	return fmt.Errorf(
		"firmware version %s not found",
		version,
	)
}
