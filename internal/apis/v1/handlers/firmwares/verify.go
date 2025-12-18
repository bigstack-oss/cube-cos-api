package firmwares

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) isFirmwareExists() bool {
	segments := strings.Split(h.file, " ")
	version := segments[3]

	list, err := h.listFirmwares()
	if err != nil {
		log.Errorf("firmwares(%s): failed to list firmwares (%v)", h.reqId, err)
		return false
	}

	for _, firmware := range list.Firmwares {
		if strings.Contains(firmware.Version, version) {
			return true
		}
	}

	return false
}

func (h *helper) isFirmwareInstalled() bool {
	segments := strings.Split(h.file, " ")
	version := segments[3]

	list, err := h.listFirmwares()
	if err != nil {
		log.Errorf("firmwares(%s): failed to list firmwares (%v)", h.reqId, err)
		return false
	}

	for _, firmware := range list.Firmwares {
		if !strings.Contains(firmware.Version, version) {
			continue
		}

		if firmware.Status.Current == status.Available {
			return false
		} else {
			return true
		}
	}

	return false
}

func (h *helper) checkFirmwareDuplication() (bool, error) {
	entries, err := os.ReadDir(firmwares.UpdateDir)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read firmware upload directory %s(%v)", h.reqId, firmwares.UpdateDir, err)
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if entry.Name() == h.file {
			return true, nil
		}
	}

	return false, nil
}

func (h *helper) checkFirmwarePattern() error {
	segments := strings.Split(h.file, " ")
	if len(segments) < 3 {
		return fmt.Errorf(
			"invalid firmware version format: %s, expected format: CUBE Appliance <version>",
			h.file,
		)
	}

	return nil
}

func (h *helper) findMatchedPkg(version string) (string, error) {
	entries, err := os.ReadDir(firmwares.UpdateDir)
	if err != nil {
		log.Errorf("firmwares(%s): failed to read update directory %s(%v)", h.reqId, firmwares.UpdateDir, err)
		return "", err
	}

	segments := strings.Split(version, " ")
	if len(segments) < 3 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	version = segments[2]
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		file := filepath.Join(firmwares.UpdateDir, entry.Name())
		if !strings.HasSuffix(file, ".pkg") {
			continue
		}

		pkgPrefix := fmt.Sprintf("_%s_", version)
		if !strings.Contains(file, pkgPrefix) {
			continue
		}

		return file, nil
	}

	return "", fmt.Errorf(
		"firmware version %s not found",
		version,
	)
}

func (h *helper) verifyFirmwareAndMd5() (*integrityResult, error) {
	result, err := h.parseMd5Data()
	if err != nil {
		return result, err
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

	err := h.moveFile(srcPath, dstPath)
	if err != nil {
		log.Errorf("firmwares(%s): failed to move firmware file from %s to %s (%v)", h.reqId, srcPath, dstPath, err)
		return err
	}

	return nil
}

func (h *helper) moveFile(srcPath, dstPath string) error {
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

func (h *helper) checkConditionForContinue() error {
	upgrade, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get firmware upgrade progress (%v)", h.reqId, err)
		return err
	}

	for _, progress := range upgrade.Progresses {
		if progress.Status.Current == status.Failed {
			return nil
		}
	}

	return fmt.Errorf(
		"no interrupted firmware upgrade found to continue",
	)
}
