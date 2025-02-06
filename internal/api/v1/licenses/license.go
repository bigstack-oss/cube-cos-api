package licenses

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	licenseExtension = ".license"
	licenseStorePath = "/var/support"
)

func getLicenseStorePath(filename string) (string, error) {
	filePath, err := filepath.Abs(filepath.Join(licenseStorePath, filename))
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(filePath, licenseStorePath) {
		return "", errors.New("invalid filename")
	}

	return filePath, nil
}
