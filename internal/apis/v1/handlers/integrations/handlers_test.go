package integrations

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

// A storage Manila still uses is refused before the 202, with the reason; a
// forwarded request is not checked again, its node refuses it while deleting.
func TestDeleteStorageRefusesWhatManilaStillUses(t *testing.T) {
	savedExists, savedManila, savedPeers, savedRecord, savedHa := storageExists, storageManilaFree, peerControls, recordStorageReq, base.IsHaEnabled
	t.Cleanup(func() {
		storageExists, storageManilaFree, peerControls, recordStorageReq, base.IsHaEnabled = savedExists, savedManila, savedPeers, savedRecord, savedHa
	})
	storageExists = func(string) (bool, error) { return true, nil }
	peerControls = func() ([]nodes.Node, error) { return nil, nil }
	recordStorageReq = func(*helper) {}
	base.IsHaEnabled = false
	inUse := "storage nfs1 is still used by Manila, delete these first: shares s1 (error)"

	tests := []struct {
		name       string
		reason     string
		forwarded  bool
		wantCode   int
		wantMsg    string
		wantChecks int
		wantQueued int
	}{
		{"in use", inUse, false, http.StatusConflict, inUse, 1, 0},
		{"free", "", false, http.StatusAccepted, "", 1, 1},
		{"forwarded", inUse, true, http.StatusAccepted, "", 0, 1},
	}

	for _, tt := range tests {
		checks := 0
		storageManilaFree = func(name string) error {
			checks++
			require.Equal(t, "nfs1", name)
			if tt.reason == "" {
				return nil
			}
			return errors.New(tt.reason)
		}

		router := gin.New()
		router.DELETE("/integrations/storages/:storageName", deleteStorage)
		req := httptest.NewRequest(http.MethodDelete, "/integrations/storages/nfs1", nil)
		if tt.forwarded {
			req.Header.Set(nodes.ForwardedHeader, "cube46")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, tt.wantCode, w.Code, tt.name)
		if tt.wantMsg != "" {
			require.Equal(t, tt.wantMsg, json.Get(w.Body.Bytes(), "msg").ToString(), tt.name)
		}
		require.Equal(t, tt.wantChecks, checks, tt.name)
		require.Equal(t, tt.wantQueued, reqQueue.Len(), tt.name)
		for reqQueue.Len() > 0 {
			queued, _ := reqQueue.Get()
			reqQueue.Done(queued)
		}
	}
}
