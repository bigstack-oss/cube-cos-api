package firmwares

import (
	"os"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) resetTmpSpace() error {
	err := os.RemoveAll(firmwares.TmpUploadDir)
	if err != nil {
		log.Errorf("firmwares(%s): failed to clean up tmp upload directory %s(%v)", h.reqId, firmwares.TmpUploadDir, err)
		return err
	}

	err = os.MkdirAll(firmwares.TmpUploadDir, 0755)
	if err != nil {
		log.Errorf("firmwares(%s): failed to recreate tmp upload directory %s(%v)", h.reqId, firmwares.TmpUploadDir, err)
		return err
	}

	return nil
}
