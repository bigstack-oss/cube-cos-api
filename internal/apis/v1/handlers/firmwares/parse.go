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
		return h.parseUploadParams()
	default:
		return nil
	}
}

func (h *helper) parseUploadParams() error {
	h.file = h.c.DefaultQuery("file", "")
	if h.file == "" {
		return fmt.Errorf("file parameter is required")
	}

	return nil
}

func (h *helper) saveUploadFirmware() error {
	path := filepath.Join(firmwares.TmpUploadDir, h.file)
	out, err := os.Create(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create firmware file %s(%v)", h.reqId, path, err)
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, h.c.Request.Body)
	if err != nil {
		log.Errorf("firmwares(%s): failed to do firmware streaming copy %s(%v)", h.reqId, path, err)
		return err
	}

	return nil
}
