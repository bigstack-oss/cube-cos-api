package fixpacks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	log "go-micro.dev/v5/logger"
)

func (h *helper) resetTmpFixpackArtifacts() error {
	path := filepath.Join(fixpacks.TmpUploadDir, fixpacks.TmpPreCalculateMd5)
	err := os.Remove(path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to clean up precalculated m5d %s(%v)", h.reqId, path, err)
		return err
	}

	entries, err := os.ReadDir(fixpacks.TmpUploadDir)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to read tmp upload directory %s(%v)", h.reqId, fixpacks.TmpUploadDir, err)
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(fixpacks.TmpUploadDir, entry.Name())
		if !strings.HasSuffix(file, ".fixpack") {
			continue
		}

		err = os.Remove(file)
		if err != nil {
			log.Errorf("fixpacks(%s): failed to remove tmp fixpack file %s(%v)", h.reqId, file, err)
			return err
		}
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
