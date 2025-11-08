package firmwares

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listFirmwares":
		return h.parseListParams()
	case "uploadFirmware":
		return h.parseUploadFirmwareParams()
	case "updateFirmware":
		return h.parseUpdateParams()
	case "uploadFirmwareMd5Sum":
		return h.parseUploadMd5Params()
	case "verfiyFirmwareAndMd5Sum":
		return h.parseVerificationParams()
	case "parseUpdateInterruptedParams":
		return h.parseUpdateInterruptedParams()
	case "deleteFirmware":
		return h.parseDeleteParams()
	case "getFirmwareNodeResolvedStatus":
		return h.parseGetNodeResolvedStatusParams()
	case "updateFirmwareTask":
		return h.parseUpdateFirmwareTaskParams()
	case "continueInterruptedFirmwareUpdate":
		return h.parseContinueInterruptionParams()
	case "retryNodeFirmwareUpdate":
		return h.parseRetryNodeFirmwareUpdateParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	if err != nil {
		log.Errorf("firmwares(%s): failed to get page parameters (%v)", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) parseUploadFirmwareParams() error {
	h.file = h.c.DefaultQuery("file", "")
	if h.file == "" {
		return fmt.Errorf("file parameter is required")
	}

	return nil
}

func (h *helper) parseUpdateParams() error {
	err := h.c.ShouldBindJSON(&h.reqOpts)
	if err != nil {
		log.Errorf("firmwares(%s): failed to bind update request options (%v)", h.reqId, err)
		return err
	}

	if h.reqOpts.Version == "" {
		err := fmt.Errorf("version is required for firmware update")
		log.Errorf("firmwares(%s): %v", h.reqId, err)
		return err
	}

	h.reqOpts.PkgPath, err = h.findMatchedPkg(h.reqOpts.Version)
	if err != nil {
		log.Errorf("firmwares(%s): failed to find matched package for version %s (%v)", h.reqId, h.reqOpts.Version, err)
		return err
	}

	h.reqOpts.SetInstalling()
	return nil
}

func (h *helper) parseRetryNodeFirmwareUpdateParams() error {
	h.reqOpts.Hostname = h.c.Param("nodeName")
	if h.reqOpts.Hostname == "" {
		return fmt.Errorf("nodeName parameter is required")
	}

	h.reqOpts.Version = h.c.Param("version")
	if h.reqOpts.Version == "" {
		return fmt.Errorf("version parameter is required")
	}

	var err error
	h.reqOpts.PkgPath, err = h.findMatchedPkg(h.reqOpts.Version)
	if err != nil {
		log.Errorf("firmwares(%s): failed to find matched package for version %s (%v)", h.reqId, h.reqOpts.Version, err)
		return err
	}

	progress, err := h.getFirmwareUpgradeProgress()
	if err != nil {
		log.Errorf("firmwares(%s): failed to get firmware upgrade progress (%v)", h.reqId, err)
		return err
	}

	h.reqOpts.AutoRolling = progress.IsRollingApplied
	h.reqOpts.SetInstalling()
	return nil
}

func (h *helper) parseUploadMd5Params() error {
	h.file = firmwares.DefaultMd5File
	return nil
}

func (h *helper) parseVerificationParams() error {
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
		if strings.HasSuffix(file, ".pkg") {
			h.file = entry.Name()
			return nil
		}
	}

	return fmt.Errorf(
		"no firmware file found in %s",
		firmwares.TmpUploadDir,
	)
}

func (h *helper) parseUpdateInterruptedParams() error {
	h.reqOpts.Hostname = h.c.Param("nodeName")
	if h.reqOpts.Hostname == "" {
		return fmt.Errorf("nodeName parameter is required")
	}

	return h.parseListParams()
}

func (h *helper) parseDeleteParams() error {
	h.file = h.c.Param("version")
	if h.file == "" {
		return fmt.Errorf("version parameter is required")
	}

	err := h.parseListParams()
	if err != nil {
		return err
	}

	return h.checkFirmwarePattern()
}

func (h *helper) parseGetNodeResolvedStatusParams() error {
	h.reqOpts.Hostname = h.c.Param("nodeName")
	if h.reqOpts.Hostname == "" {
		return fmt.Errorf("node name parameter is required")
	}

	return nil
}

func (h *helper) parseUpdateFirmwareTaskParams() error {
	err := h.c.ShouldBindJSON(&h.reqOpts)
	if err != nil {
		log.Errorf("firmwares(%s): failed to bind update request options (%v)", h.reqId, err)
		return err
	}

	if h.reqOpts.Version == "" {
		err := fmt.Errorf("version is required for firmware update task")
		log.Errorf("firmwares(%s): %v", h.reqId, err)
		return err
	}

	if h.reqOpts.Hostname == "" {
		err := fmt.Errorf("hostname is required for firmware update task")
		log.Errorf("firmwares(%s): %v", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) parseContinueInterruptionParams() error {
	h.reqOpts.Hostname = h.c.Param("nodeName")
	if h.reqOpts.Hostname == "" {
		return fmt.Errorf("nodeName parameter is required")
	}

	return nil
}

func (h *helper) saveUploadFile() error {
	path := filepath.Join(firmwares.TmpUploadDir, h.file)
	out, err := os.Create(path)
	if err != nil {
		log.Errorf("firmwares(%s): failed to create %s %s(%v)", path, h.reqId, path, err)
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, h.c.Request.Body)
	if err != nil {
		log.Errorf("firmwares(%s): failed to do %s streaming copy %s(%v)", path, h.reqId, path, err)
		return err
	}

	return nil
}
