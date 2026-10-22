package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// PathIn returns the path of the file name inside dir. name must be a single
// path element, so the result is always a direct child of dir.
func PathIn(dir, name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.Contains(name, "/") {
		return "", fmt.Errorf("invalid file name %q", name)
	}

	return filepath.Join(dir, name), nil
}

// Save stores an uploaded file without changing the mode of the destination
// directory. gin's SaveUploadedFile chmods filepath.Dir(dst) on every save.
func Save(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
