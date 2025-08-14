package firmwares

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) verifyFirmwareAndMd5() (*integrityResult, error) {
	prePath := filepath.Join(firmwares.TmpUploadDir, firmwares.TmpPreCalculateMd5)
	precalculated, err := os.ReadFile(prePath)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read precalculated md5 %s(%v)", h.reqId, prePath, err)
		return nil, err
	}

	defaultPath := filepath.Join(firmwares.TmpUploadDir, firmwares.DefaultMd5File)
	expected, err := os.ReadFile(defaultPath)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read md5 file %s(%v)", h.reqId, defaultPath, err)
		return nil, err
	}

	result := &integrityResult{
		FirmwareMd5: strings.Trim(string(precalculated), "\n"),
		ExpectedMd5: strings.Trim(string(expected), "\n"),
	}

	if !strings.Contains(result.ExpectedMd5, result.FirmwareMd5) {
		return result, fmt.Errorf(
			"md5 verification failed: expected %s, got %s",
			string(result.ExpectedMd5),
			string(result.FirmwareMd5),
		)
	}

	return result, nil
}

func (h *helper) setValidFirmware() error {
	srcPath := filepath.Join(firmwares.TmpUploadDir, h.file)
	dstPath := filepath.Join(firmwares.UpdateDir, h.file)

	err := h.MoveFile(srcPath, dstPath)
	if err != nil {
		log.Errorf("firmwares(%s): failed to move firmware file from %s to %s (%v)", h.reqId, srcPath, dstPath, err)
		return err
	}

	return nil
}

func (h *helper) MoveFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Errorf("firmwares(%s): failed to open source file %s (%v)", h.reqId, srcPath, err)
		return err
	}

	defer srcFile.Close()
	dstFile, err := os.Create(dstPath)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create destination file %s (%v)", h.reqId, dstPath, err)
		return err
	}

	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	srcFile.Close()
	if err != nil {
		log.Errorf("firmwares(%s): failed to copy file from %s to %s (%v)", h.reqId, srcPath, dstPath, err)
		return err
	}

	return nil
}
