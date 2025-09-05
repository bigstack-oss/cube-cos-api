package firmwares

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) resetTmpFirmwareArtifacts() error {
	err := h.syncTmpUploadDir()
	if err != nil {
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
		if !strings.HasSuffix(file, ".pkg") && entry.Name() != firmwares.TmpPreCalculateMd5 {
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

func (h *helper) syncTmpUploadDir() error {
	_, err := os.Stat(firmwares.TmpUploadDir)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(firmwares.TmpUploadDir, 0755)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create tmp upload directory %s(%v)", h.reqId, firmwares.TmpUploadDir, err)
		return err
	}

	return nil
}

func (h *helper) resetTmpFirmwareMd5() error {
	path := filepath.Join(firmwares.TmpUploadDir, firmwares.DefaultMd5File)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to reset tmp firmware m5d %s(%v)", h.reqId, path, err)
		return err
	}

	return nil
}
