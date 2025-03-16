package cubecos

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	log "go-micro.dev/v5/logger"
)

func CreateSupportFile(supportFile v1.SupportFile) error {
	path, err := CreateSupportCommentFile(supportFile.Comment)
	if err != nil {
		log.Errorf("supportFile: failed to create support comment file: %s", err.Error())
		return err
	}

	out, err := exec.Command("hex_config", "create_support_file", path).CombinedOutput()
	if err != nil {
		log.Errorf("supportFile: failed to create support file: %s", string(out))
		return err
	}

	return nil
}

func CreateSupportCommentFile(comment string) (string, error) {
	randomSize := make([]byte, 8)
	_, err := rand.Read(randomSize)
	if err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomSize)
	filePath := filepath.Join("/tmp/support-comment-file", randomStr)
	err = os.WriteFile(filePath, []byte(comment), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func GetSupportFileByComment(comment string) (string, error) {
	files, err := os.ReadDir(v1.DefaultSupportFileDir)
	if err != nil {
		log.Errorf("supportFile: failed to read support file directory: %s", err.Error())
		return "", err
	}

	if len(files) == 0 {
		err := errors.New("no support file found")
		log.Errorf("supportFile: %v", err)
		return "", err
	}

	file, err := findSupportFile(files, comment)
	if err != nil {
		log.Errorf("supportFile: %v", err)
		return "", err
	}

	return filepath.Join("/var/support", file.Name()), nil
}

func findSupportFile(files []os.DirEntry, comment string) (os.DirEntry, error) {
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !isSupportFile(file.Name()) {
			continue
		}

		content, err := os.ReadFile(filepath.Join(v1.DefaultSupportFileDir, file.Name()))
		if err != nil {
			continue
		}

		if string(content) == comment {
			return file, nil
		}
	}

	return nil, errors.New("no support file found")
}

func isSupportFile(file string) bool {
	return strings.HasPrefix(file, fmt.Sprintf("CUBE_%s", v1.DataCenterVersion)) &&
		strings.HasSuffix(file, fmt.Sprintf("%s.support", v1.Hostname))
}
