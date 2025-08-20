package fixpacks

import (
	"os"
	"path/filepath"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	log "go-micro.dev/v5/logger"
)

func (h *helper) resetTmpFixpackArtifacts() error {
	os.RemoveAll(fixpacks.TmpUploadDir)
	err := os.MkdirAll(fixpacks.TmpUploadDir, 0755)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to reset tmp fixpack artifacts dir(%v)", h.reqId, err)
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
