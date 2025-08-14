package firmwares

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listFirmwares":
		return h.parseListParams()
	case "uploadFirmware":
		return h.parseUploadFirmwareParams()
	case "uploadFirmwareMd5Sum":
		return h.parseUploadMd5Params()
	case "verfiyFirmwareAndMd5Sum":
		return h.parseVerificationParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get page parameters (%v)", h.reqId, err)
		return err
	}

	return nil
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

func (h *helper) parseVerificationParams() error {
	entries, err := os.ReadDir(firmwares.TmpUploadDir)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read tmp upload directory %s(%v)", h.reqId, firmwares.TmpUploadDir, err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(firmwares.TmpUploadDir, entry.Name())
		if strings.HasSuffix(file, ".pkg") {
			h.file = entry.Name()
			return nil
		}
	}

	return fmt.Errorf("no firmware file found in %s", firmwares.TmpUploadDir)
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
