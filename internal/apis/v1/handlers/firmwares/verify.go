package firmwares

import (
	"fmt"
	"os"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) verifyFirmwareAndMd5() (*integrityResult, error) {
	precalculated, err := os.ReadFile(firmwares.TmpPreCalculateMd5)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read precalculated md5 %s(%v)", h.reqId, firmwares.TmpPreCalculateMd5, err)
		return nil, err
	}

	expected, err := os.ReadFile(firmwares.DefaultMd5File)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read md5 file %s(%v)", h.reqId, firmwares.DefaultMd5File, err)
		return nil, err
	}

	result := &integrityResult{
		FirmwareMd5: string(precalculated),
		ExpectedMd5: string(expected),
	}

	if !strings.Contains(result.ExpectedMd5, result.FirmwareMd5) {
		return nil, fmt.Errorf(
			"md5 verification failed: expected %s, got %s",
			string(result.ExpectedMd5),
			string(result.FirmwareMd5),
		)
	}

	return result, nil
}
