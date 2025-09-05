package fixpacks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/ceph"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	log "go-micro.dev/v5/logger"
)

func (h *helper) getVersionStatus(version string) (string, error) {
	fixpacks, err := h.listFixpacks()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to list fixpack for checking installation(%v)", h.reqId, err)
		return "", err
	}

	for _, fixpack := range fixpacks.Fixpacks {
		if fixpack.Version != version {
			continue
		}

		if fixpack.Status.Current == status.Installed {
			return fixpack.Status.Current, nil
		}
	}

	return "", fmt.Errorf(
		"fixpack version %s not found",
		version,
	)
}

func (h *helper) checkRebootRequirement() (bool, error) {
	update, err := h.getFixpackUpdateProgress()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to get fixpack update progress for checking reboot requirement(%v)", h.reqId, err)
		return false, err
	}

	fixpack, found := cubecos.GetFixpackByVersion(update.Version)
	if !found {
		return false, fmt.Errorf("fixpack version %s not found", update.Version)
	}

	return fixpack.RebootRequired, nil
}

func (h *helper) checkRollback(version string) error {
	fixpack, found := cubecos.GetFixpackByVersion(version)
	if !found {
		return fmt.Errorf("fixpack version %s not found", version)
	}

	if !fixpack.Status.IsRollbackable {
		return fmt.Errorf("fixpack version %s is not rollbackable", version)
	}

	return nil
}

func (h *helper) checkEnvConditions() error {
	if cubecos.IsInStrictMode() {
		return fmt.Errorf("env is in the strict mode, cannot proceed with fixpack operations")
	}

	if !ceph.IsHealthy() {
		return fmt.Errorf("ceph is not healthy, cannot proceed with fixpack operations")
	}

	return nil
}

func (h *helper) checkFixpackPattern() error {
	if strings.HasSuffix(h.file, ".fixpack") {
		return nil
	}

	return fmt.Errorf(
		"invalid fixpack file format: %s, expected .fixpack",
		h.file,
	)
}

func (h *helper) verifyFixpackAndMd5() (*integrityResult, error) {
	result, err := h.parseMd5Data()
	if err != nil {
		return result, err
	}

	if !strings.Contains(result.ExpectedMd5, result.FixpackMd5) {
		return result, fmt.Errorf(
			"md5 verification failed: expected %s, got %s",
			string(result.ExpectedMd5),
			string(result.FixpackMd5),
		)
	}

	return result, nil
}

func (h *helper) setValidFixpack() error {
	srcPath := filepath.Join(fixpacks.TmpUploadDir, h.file)
	dstPath := filepath.Join(fixpacks.UpdateDir, h.file)

	err := h.moveFile(srcPath, dstPath)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to move fixpack file from %s to %s (%v)", h.reqId, srcPath, dstPath, err)
		return err
	}

	return nil
}

func (h *helper) moveFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to open source file %s (%v)", h.reqId, srcPath, err)
		return err
	}

	defer srcFile.Close()
	dstFile, err := os.Create(dstPath)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to create destination file %s (%v)", h.reqId, dstPath, err)
		return err
	}

	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	srcFile.Close()
	if err != nil {
		log.Errorf("fixpacks(%s): failed to copy file from %s to %s (%v)", h.reqId, srcPath, dstPath, err)
		return err
	}

	return nil
}
