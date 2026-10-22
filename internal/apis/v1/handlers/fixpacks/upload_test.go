package fixpacks

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
	return &helper{c: c, handler: "uploadFixpack"}
}

func TestUploadParamsAcceptFileName(t *testing.T) {
	h := newUploadHelper(t, "/fixpacks?file=3.2.0-1.fixpack")
	if err := h.parseUploadFixpackParams(); err != nil {
		t.Fatalf("parseUploadFixpackParams: %v", err)
	}
	if want := "3.2.0-1.fixpack"; h.file != want {
		t.Fatalf("file = %q, want %q", h.file, want)
	}
}

func TestUploadParamsRejectPath(t *testing.T) {
	for _, file := range []string{"..", "..%2Fx.fixpack", "..%2F..%2Fetc%2Fx", "a%2Fb.fixpack", "%2Fetc%2Fx"} {
		h := newUploadHelper(t, "/fixpacks?file="+file)
		if err := h.parseUploadFixpackParams(); err == nil {
			t.Fatalf("file=%s: parseUploadFixpackParams accepted %q", file, h.file)
		}
	}
}

func TestSaveUploadFileRejectsPath(t *testing.T) {
	h := newUploadHelper(t, "/fixpacks")
	h.file = "../x.fixpack"
	if err := h.saveUploadFile(); err == nil {
		t.Fatal("saveUploadFile accepted a file name with a path")
	}
}
