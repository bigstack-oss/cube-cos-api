package upload

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// gin >= v1.12 chmods the destination's parent directory on every save.
func TestSaveKeepsDirMode(t *testing.T) {
	dir := t.TempDir()
	want := os.FileMode(0o777) | os.ModeSticky
	if err := os.Chmod(dir, want); err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", "storage-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("backends: []\n")); err != nil {
		t.Fatal(err)
	}
	w.Close()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	c.Request = req

	fh, err := c.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "storage-model.yaml")
	if err := Save(fh, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("uploaded file not written: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode() & (os.ModePerm | os.ModeSticky); got != want {
		t.Fatalf("upload changed the destination directory mode: got %v, want %v", got, want)
	}
}

func TestPathInAcceptsFileName(t *testing.T) {
	for _, name := range []string{
		"CUBE Appliance 3.2.0 x86_64.pkg",
		"3.2.0-1.fixpack",
		"notify.sh",
		"..notify.sh",
	} {
		got, err := PathIn("/tmp/firmwares", name)
		if err != nil {
			t.Fatalf("PathIn(%q): %v", name, err)
		}
		if want := "/tmp/firmwares/" + name; got != want {
			t.Fatalf("PathIn(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestPathInRejectsPath(t *testing.T) {
	for _, name := range []string{
		"",
		".",
		"..",
		"../x",
		"../../etc/x",
		"a/b",
		"/etc/x",
		"x/",
		"/",
	} {
		if got, err := PathIn("/tmp/firmwares", name); err == nil {
			t.Fatalf("PathIn(%q) = %q, want an error", name, got)
		}
	}
}

// A directory Save creates must not be world-accessible.
func TestSaveCreatesDirNotWorldReadable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cos-storages")

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("file", "storage-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("backends: []\n")); err != nil {
		t.Fatal(err)
	}
	w.Close()

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	c.Request = req

	fh, err := c.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(fh, filepath.Join(dir, "storage-model.yaml")); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got&0o007 != 0 {
		t.Fatalf("created directory is world-accessible: %04o", got)
	}
}
