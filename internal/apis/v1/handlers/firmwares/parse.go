package firmwares

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "uploadFirmware":
		return h.parseUploadFirmwareParams()
	case "uploadFirmwareMd5Sum":
		return h.parseUploadMd5Params()
	default:
		return nil
	}
}

func (h *helper) parseUploadFirmwareParams() error {
	h.file = h.c.DefaultQuery("file", "")
	if h.file == "" {
		return fmt.Errorf("file parameter is required")
	}

	return nil
}

func (h *helper) parseUploadMd5Params() error {
	h.file = firmwares.DefaultMd5File
	return nil
}

func (h *helper) saveUploadFile() error {
	path := filepath.Join(firmwares.TmpUploadDir, h.file)
	out, err := os.Create(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create %s %s(%v)", path, h.reqId, path, err)
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, h.c.Request.Body)
	if err != nil {
		log.Errorf("firmwares(%s): failed to do %s streaming copy %s(%v)", path, h.reqId, path, err)
		return err
	}

	return nil
}
