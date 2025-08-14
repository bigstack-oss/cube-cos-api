package firmwares

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) resetTmpFirmwareArtifacts() error {
	err := os.Remove(firmwares.TmpPreCalculateMd5)
	if err != nil {
		log.Errorf("firmwares(%s): failed to clean up precalculated m5d %s(%v)", h.reqId, firmwares.TmpPreCalculateMd5, err)
		return err
	}

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
		if !strings.HasSuffix(file, ".pkg") {
			continue
		}

		err = os.Remove(file)
		if err != nil {
			log.Errorf("firmwares(%s): failed to remove tmp firmware file %s(%v)", h.reqId, file, err)
			return err
		}
	}

	return nil
}

func (h *helper) resetTmpFirmwareMd5() error {
	err := os.Remove(firmwares.DefaultMd5File)
	if err != nil {
		log.Errorf("firmwares(%s): failed to reset tmp firmware m5d %s(%v)", h.reqId, firmwares.DefaultMd5File, err)
		return err
	}

	return nil
}
