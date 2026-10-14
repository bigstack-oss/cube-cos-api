package integrations

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
)

func TestIsPeerRequest(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    bool
	}{
		{"user request through the VIP", nil, false},
		{"CLI request with the node token", map[string]string{"Node": "cube46", "Authorization": "Bearer t"}, false},
		{"request forwarded by a control", map[string]string{"Node": "cube46", nodes.ForwardedHeader: "cube46"}, true},
	}

	for _, tt := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/api/v1/datacenters/dc/integrations/storages", nil)
		for k, v := range tt.headers {
			c.Request.Header.Set(k, v)
		}

		h := &helper{c: c}
		if got := h.isPeerRequest(); got != tt.want {
			t.Errorf("%s: isPeerRequest() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestForwardedRequestIsMarked(t *testing.T) {
	base.Hostname = "cube45"
	h := &helper{}
	in := http.Header{}
	in.Set("Node", "cube46")
	in.Set("Authorization", "Bearer t")

	got := h.convertHeadersToMap(in)
	if got[nodes.ForwardedHeader] != "cube45" {
		t.Fatalf("forwarded headers lack %s: %v", nodes.ForwardedHeader, got)
	}
}
