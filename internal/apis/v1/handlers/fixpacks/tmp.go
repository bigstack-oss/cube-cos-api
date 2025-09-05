package fixpacks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	log "go-micro.dev/v5/logger"
)

func (h *helper) resetTmpFixpackArtifacts() error {
	err := h.syncTmpUploadDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(fixpacks.TmpUploadDir)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to read tmp upload directory %s(%v)", h.reqId, firmwares.TmpUploadDir, err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.TmpUploadDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") && entry.Name() != fixpacks.TmpPreCalculateMd5 {
			continue
		}

		err = os.Remove(file)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to remove tmp fixpack %s(%v)", h.reqId, file, err)
			return err
		}
	}

	return nil
}

func (h *helper) syncTmpUploadDir() error {
	_, err := os.Stat(fixpacks.TmpUploadDir)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(fixpacks.TmpUploadDir, 0755)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to create tmp upload directory %s(%v)", h.reqId, fixpacks.TmpUploadDir, err)
		return err
	}

	return nil
}

func (h *helper) resetTmpFixpackMd5() error {
	path := filepath.Join(fixpacks.TmpUploadDir, fixpacks.DefaultMd5File)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to reset tmp fixpack m5d %s(%v)", h.reqId, path, err)
		return err
	}

	return nil
}
