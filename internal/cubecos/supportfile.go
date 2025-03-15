package cubecos

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "go-micro.dev/v5/logger"
)

func CreateSupportFile() error {
	out, err := exec.Command("hex_config", "create_support_file").CombinedOutput()
	if err != nil {
		log.Errorf("supportFile: failed to create support file: %s", string(out))
		return err
	}

	return nil
}

func GetNewSupportFile() (string, error) {
	files, err := os.ReadDir("/var/support")
	if err != nil {
		log.Errorf("supportFile: failed to read support file directory: %s", err.Error())
		return "", err
	}

	if len(files) == 0 {
		err := errors.New("no support file found")
		log.Errorf("supportFile: %v", err)
		return "", err
	}

	return findLastFile(files)
}

func findLastFile(files []os.DirEntry) (string, error) {
	fileName := ""
	fileTime := time.Time{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(fileTime) {
			fileTime = info.ModTime()
			fileName = file.Name()
		}
	}

	if fileName == "" {
		return "", errors.New("no support file found")
	}

	return filepath.Join("/var/support", fileName), nil
}
