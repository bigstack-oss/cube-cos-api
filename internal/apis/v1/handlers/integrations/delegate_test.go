package integrations

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	httphelper "github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
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

// storagePeer answers like a control's storage handlers; fail answers 500
func storagePeer(t *testing.T, name string, fail bool, mu *sync.Mutex, got *[]string) nodes.Node {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		*got = append(*got, name+" "+req.Method+" "+req.Header.Get(nodes.ForwardedHeader))
		mu.Unlock()
		if fail {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)
	return nodes.Node{Hostname: name, Protocol: "http", Address: strings.TrimPrefix(srv.URL, "http://")}
}

func TestUpdateStorageMarksOriginWithAcceptingPeers(t *testing.T) {
	savedPeers, savedRecord, savedHa := peerControls, recordStorageReq, base.IsHaEnabled
	t.Cleanup(func() { peerControls, recordStorageReq, base.IsHaEnabled = savedPeers, savedRecord, savedHa })
	recordStorageReq = func(*helper) {}
	base.Hostname = "cube45"
	client, _ := httphelper.NewHelper(httphelper.RetryCount(0))

	tests := []struct {
		name       string
		handler    string
		method     string
		desired    string
		forwarded  bool
		ha         bool
		wantOrigin bool
		wantPeers  string
		wantSent   int
	}{
		{"create through the VIP", "createStorage", http.MethodPost, status.Created, false, true, true, "cube46", 2},
		{"update from the CLI", "updateStorage", http.MethodPatch, status.Updated, false, true, true, "cube46", 2},
		{"delete through the VIP", "deleteStorage", http.MethodDelete, status.Deleted, false, true, true, "cube46", 2},
		{"forwarded create", "createStorage", http.MethodPost, status.Created, true, true, false, "", 0},
		{"set as default", "setStorageAsDefault", http.MethodPatch, status.Defaulted, false, true, false, "", 2},
		{"single node", "createStorage", http.MethodPost, status.Created, false, false, false, "", 0},
	}

	for _, tt := range tests {
		var mu sync.Mutex
		sent := []string{}
		peers := []nodes.Node{storagePeer(t, "cube46", false, &mu, &sent), storagePeer(t, "cube47", true, &mu, &sent)}
		base.IsHaEnabled = tt.ha
		peerControls = func() ([]nodes.Node, error) {
			if !base.IsHaEnabled {
				return nil, nil
			}
			return peers, nil
		}

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(tt.method, "/api/v1/datacenters/dc/integrations/storages", strings.NewReader(`{"name":"pure1"}`))
		if tt.forwarded {
			c.Request.Header.Set(nodes.ForwardedHeader, "cube46")
		}
		h := &helper{c: c, handler: tt.handler, http: client, rawBody: []byte(`{"name":"pure1"}`)}
		h.storageReqOpts.CinderDetails.Name = "pure1"
		h.storageReqOpts.Status.Desired = tt.desired

		h.updateStorage()
		queued, _ := reqQueue.Get()
		reqQueue.Done(queued)

		if queued.IsOrigin != tt.wantOrigin || strings.Join(queued.Peers, ",") != tt.wantPeers {
			t.Errorf("%s: queued origin=%v peers=%q, want origin=%v peers=%q",
				tt.name, queued.IsOrigin, queued.Peers, tt.wantOrigin, tt.wantPeers)
		}
		if len(sent) != tt.wantSent {
			t.Errorf("%s: forwarded %d times (%q), want %d", tt.name, len(sent), sent, tt.wantSent)
		}
		for _, s := range sent {
			if !strings.HasSuffix(s, " cube45") {
				t.Errorf("%s: forward not marked: %q", tt.name, s)
			}
		}
	}
}
