package md5

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GenByFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s(%v)", file, err)
	}

	defer f.Close()
	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("failed to calculate md5 for file %s(%v)", file, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
