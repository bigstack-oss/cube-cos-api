package firmwares

import (
	"os"
	"path"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/md5"
	log "go-micro.dev/v5/logger"
)

func (h *helper) syncFirmwareMd5() error {
	path := path.Join(firmwares.TmpUploadDir, h.file)
	sum, err := md5.GenByFile(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to generate md5 sum for firmware file %s(%v)", h.reqId, path, err)
		return err
	}

	err = os.WriteFile(firmwares.TmpPreCalculateMd5, []byte(sum), 0644)
	if err != nil {
		log.Errorf("firmwares(%s): failed to write md5 sum to file %s(%v)", h.reqId, firmwares.TmpPreCalculateMd5, err)
		return err
	}

	return nil
}
