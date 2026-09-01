package upload

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

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
