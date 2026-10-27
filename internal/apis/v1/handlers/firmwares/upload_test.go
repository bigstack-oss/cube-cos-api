package firmwares

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/firmwares"
	"github.com/gin-gonic/gin"
)

func newUploadHelper(t *testing.T, target string) *helper {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", target, nil)
	return &helper{c: c, handler: "uploadFirmware"}
}

func TestUploadParamsAcceptFileName(t *testing.T) {
	h := newUploadHelper(t, "/firmwares?file=CUBE%20Appliance%203.2.0%20x86_64.pkg")
	if err := h.parseUploadFirmwareParams(); err != nil {
		t.Fatalf("parseUploadFirmwareParams: %v", err)
	}
	if want := "CUBE Appliance 3.2.0 x86_64.pkg"; h.file != want {
		t.Fatalf("file = %q, want %q", h.file, want)
	}
	if want := filepath.Join(firmwares.TmpUploadDir, "CUBE Appliance 3.2.0 x86_64.pkg"); h.uploadPath != want {
		t.Fatalf("uploadPath = %q, want %q", h.uploadPath, want)
	}
}

func TestUploadParamsRejectPath(t *testing.T) {
	for _, file := range []string{"..", "..%2Fx.pkg", "..%2F..%2Fetc%2Fx", "a%2Fb.pkg", "%2Fetc%2Fx"} {
		h := newUploadHelper(t, "/firmwares?file="+file)
		if err := h.parseUploadFirmwareParams(); err == nil {
			t.Fatalf("file=%s: parseUploadFirmwareParams accepted %q", file, h.file)
		}
	}
}

func TestSaveUploadFileWritesUploadPath(t *testing.T) {
	h := newUploadHelper(t, "/firmwares")
	h.c.Request.Body = io.NopCloser(strings.NewReader("data"))
	h.uploadPath = filepath.Join(t.TempDir(), "x.pkg")
	if err := h.saveUploadFile(); err != nil {
		t.Fatalf("saveUploadFile: %v", err)
	}
	got, err := os.ReadFile(h.uploadPath)
	if err != nil || string(got) != "data" {
		t.Fatalf("uploadPath content = %q, %v", got, err)
	}
}
