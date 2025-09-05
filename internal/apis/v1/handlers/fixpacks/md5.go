package fixpacks

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/fixpacks"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/md5"
	log "go-micro.dev/v5/logger"
)

func (h *helper) syncFixpackMd5() error {
	path := filepath.Join(fixpacks.TmpUploadDir, h.file)
	sum, err := md5.GenByFile(path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to generate md5 sum for fixpack file %s(%v)", h.reqId, path, err)
		return err
	}

	path = filepath.Join(fixpacks.TmpUploadDir, fixpacks.TmpPreCalculateMd5)
	err = os.WriteFile(path, []byte(sum), 0644)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to write md5 sum to file %s(%v)", h.reqId, path, err)
		return err
	}

	return nil
}

func (h *helper) parseMd5Data() (*integrityResult, error) {
	path := filepath.Join(fixpacks.TmpUploadDir, fixpacks.TmpPreCalculateMd5)
	precalculated, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to read precalculated md5 %s(%v)", h.reqId, path, err)
		return nil, err
	}

	path = filepath.Join(fixpacks.TmpUploadDir, fixpacks.DefaultMd5File)
	expected, err := os.ReadFile(path)
	if err != nil {
		log.Errorf("fixpacks(%s): failed to read md5 file %s(%v)", h.reqId, path, err)
		return nil, errors.New("please upload the md5 file before verification")
	}

	return &integrityResult{
		FixpackMd5:  h.LeavePureTextOnly(string(precalculated)),
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
