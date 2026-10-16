package cubecos

import (
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/storages"
	"go.mongodb.org/mongo-driver/bson"
)

// The filter must match a stored storages.ReqOpts by its bson keys.
func TestHostProcessingFilterMatchesStoredStorageRequests(t *testing.T) {
	base.Hostname = "cube45"
	req := storages.ReqOpts{Hostname: "cube45", Status: status.Storage{IsProcessing: true}}
	raw, err := bson.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var doc bson.M
	if err := bson.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}

	f := hostProcessingFilter()
	if doc["hostname"] != f["hostname"] {
		t.Errorf("hostname: stored %v, filter %v", doc["hostname"], f["hostname"])
	}
	st, ok := doc["status"].(bson.M)
	if !ok || st["isProcessing"] != f["status.isProcessing"] {
		t.Errorf("status.isProcessing: stored %v, filter %v", doc["status"], f["status.isProcessing"])
	}
	if _, ok := doc["host"]; ok {
		t.Errorf("storage requests have no host key; the old filter keyed on it")
	}
}
