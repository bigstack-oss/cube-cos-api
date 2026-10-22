package firmwares

import (
	"net/http/httptest"
	"testing"

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
}

func TestUploadParamsRejectPath(t *testing.T) {
	for _, file := range []string{"..", "..%2Fx.pkg", "..%2F..%2Fetc%2Fx", "a%2Fb.pkg", "%2Fetc%2Fx"} {
		h := newUploadHelper(t, "/firmwares?file="+file)
		if err := h.parseUploadFirmwareParams(); err == nil {
			t.Fatalf("file=%s: parseUploadFirmwareParams accepted %q", file, h.file)
		}
	}
}

func TestSaveUploadFileRejectsPath(t *testing.T) {
	h := newUploadHelper(t, "/firmwares")
	h.file = "../x.pkg"
	if err := h.saveUploadFile(); err == nil {
		t.Fatal("saveUploadFile accepted a file name with a path")
	}
}
