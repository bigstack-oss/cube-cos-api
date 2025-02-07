package cubecos

import (
	"bufio"
	"os"
	"strings"

	"github.com/bigstack-oss/cube-cos-api/internal/config"
)

func ReadSettingSys(settingName string) (string, error) {
	sys, err := os.Open(config.Opts.Spec.Os.System)
	if err != nil {
		return "", err
	}

	defer sys.Close()
	value := ""
	scanner := bufio.NewScanner(sys)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isCommentOrBlank(line) {
			continue
		}

		if !strings.HasPrefix(line, settingName) {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if isKeyValuePattern(parts) {
			value = strings.TrimSpace(parts[1])
			break
		}
	}
	err = scanner.Err()
	if err != nil {
		return "", err
	}

	return value, nil
}

func isCommentOrBlank(line string) bool {
	return strings.HasPrefix(line, "#") || line == ""
}

func isKeyValuePattern(parts []string) bool {
	return len(parts) == 2
}
