package fixpacks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	log "go-micro.dev/v5/logger"
)

func (h *helper) parseParamsByHandler() error {
	switch h.handler {
	case "listFixpacks":
		return h.parseListParams()
	case "uploadFixpack":
		return h.parseUploadFixpackParams()
	case "uploadFixpackMd5Sum":
		return h.parseUploadMd5Params()
	case "verfiyFixpackAndMd5Sum":
		return h.parseVerificationParams()
	case "listUpdatableNodes":
		return h.parseListUpdatableParams()
	case "installFixpack":
		return h.parseInstallParams()
	case "getFixpackUpdateProgress":
		return h.parseGetProgressParams()
	case "continueInterruptedFixpackUpdate":
		return h.parseUpdateInterruptedParams()
	case "rollbackFixpack":
		return h.parseRollbackParams()
	case "deleteFixpack":
		return h.parseDeleteFixpackParams()
	case "updateFixpackTask":
		return h.parseUpdateFixpackTaskParams()
	default:
		return nil
	}
}

func (h *helper) parseListParams() error {
	var err error
	h.page, err = queries.GetPage(h.c)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get page parameters (%v)", h.reqId, err)
		return err
	}

	return nil
}

func (h *helper) parseUploadFixpackParams() error {
	h.file = h.c.DefaultQuery("file", "")
	if h.file == "" {
		return fmt.Errorf("file parameter is required")
	}

	return nil
}

func (h *helper) parseUploadMd5Params() error {
	h.file = fixpacks.DefaultMd5File
	return nil
}

func (h *helper) parseVerificationParams() error {
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
		if strings.HasSuffix(file, ".fixpack") {
			h.file = entry.Name()
			return nil
		}
	}

	return fmt.Errorf(
		"no fixpack file found in %s",
		fixpacks.TmpUploadDir,
	)
}

func (h *helper) parseListUpdatableParams() error {
	h.reqOpts.Version = h.c.Param("version")
	if h.reqOpts.Version == "" {
		return fmt.Errorf("version parameter is required")
	}

	_, found := cubecos.GetFixpackByVersion(h.reqOpts.Version)
	if !found {
		return fmt.Errorf("fixpack version %s not found", h.reqOpts.Version)
	}

	return nil
}

func (h *helper) parseInstallParams() error {
	err := h.c.ShouldBindJSON(&h.reqOpts)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to bind JSON for install parameters (%v)", h.reqId, err)
		return err
	}

	path, found := cubecos.GetFixpackPathByVersion(h.reqOpts.Version)
	if !found {
		err := fmt.Errorf("fixpack %s not found", h.reqOpts.Version)
		log.Errorf("fixpacks(%s): %v", h.reqId, err)
		return err
	}

	h.reqOpts.Path = path
	h.reqOpts.SetInstalling()
	return nil
}

func (h *helper) parseGetProgressParams() error {
	return h.parseListParams()
}

func (h *helper) parseUpdateInterruptedParams() error {
	h.reqOpts.Hostname = h.c.Param("nodeName")
	if h.reqOpts.Hostname == "" {
		return fmt.Errorf("nodeName parameter is required")
	}

	return h.parseListParams()
}

func (h *helper) parseRollbackParams() error {
	h.reqOpts.Version = h.c.Param("version")
	if h.reqOpts.Version == "" {
		return fmt.Errorf("version parameter is required")
	}

	found := false
	_, found = cubecos.GetFixpackRawByVersion(h.reqOpts.Version)
	if !found {
		return fmt.Errorf("fixpack version %s not found", h.reqOpts.Version)
	}

	err := h.checkRollback(h.reqOpts.Version)
	if err != nil {
		return err
	}

	h.reqOpts.SetRollingBack()
	return nil
}

func (h *helper) parseDeleteFixpackParams() error {
	h.reqOpts.Version = h.c.Param("version")
	if h.reqOpts.Version == "" {
		return fmt.Errorf("version parameter is required")
	}

	found := false
	h.file, found = cubecos.GetFixpackPathByVersion(h.reqOpts.Version)
	if !found {
		return fmt.Errorf("fixpack version %s not found", h.reqOpts.Version)
	}

	return nil
}

func (h *helper) parseUpdateFixpackTaskParams() error {
	err := h.c.ShouldBindJSON(&h.reqOpts)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to bind JSON for update task parameters (%v)", h.reqId, err)
		return err
	}

	if h.reqOpts.Hostname == "" {
		return fmt.Errorf("hostname parameter is required")
	}

	return nil
}

func (h *helper) saveUploadFile() error {
	path := filepath.Join(fixpacks.TmpUploadDir, h.file)
	out, err := os.Create(path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to create %s %s(%v)", path, h.reqId, path, err)
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, h.c.Request.Body)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to do %s streaming copy %s(%v)", path, h.reqId, path, err)
		return err
	}

	return nil
}
