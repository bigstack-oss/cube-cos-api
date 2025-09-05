package firmwares

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/md5"
	log "go-micro.dev/v5/logger"
)

func (h *helper) syncFirmwareMd5() error {
	path := filepath.Join(firmwares.TmpUploadDir, h.file)
	sum, err := md5.GenByFile(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to generate md5 sum for firmware file %s(%v)", h.reqId, path, err)
		return err
	}

	path = filepath.Join(firmwares.TmpUploadDir, firmwares.TmpPreCalculateMd5)
	err = os.WriteFile(path, []byte(sum), 0644)
	if err != nil {
		log.Errorf("firmwares(%s): failed to write md5 sum to file %s(%v)", h.reqId, path, err)
		return err
	}

	return nil
}

func (h *helper) parseMd5Data() (*integrityResult, error) {
	path := filepath.Join(firmwares.TmpUploadDir, firmwares.TmpPreCalculateMd5)
	precalculated, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read precalculated md5 %s(%v)", h.reqId, path, err)
		return nil, err
	}

	path = filepath.Join(firmwares.TmpUploadDir, firmwares.DefaultMd5File)
	expected, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read md5 file %s(%v)", h.reqId, path, err)
		return nil, err
	}

	return &integrityResult{
		FirmwareMd5: h.LeavePureTextOnly(string(precalculated)),
		ExpectedMd5: h.LeavePureTextOnly(string(expected)),
	}, nil
}

func (h *helper) LeavePureTextOnly(text string) string {
	return strings.NewReplacer(
		" ", "",
		"\n", "",
		"-", "",
	).Replace(text)
}
