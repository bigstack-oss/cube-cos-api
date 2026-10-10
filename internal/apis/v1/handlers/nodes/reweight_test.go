package nodes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newPatchContext(body string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c
}

// An absent reweight binds as 0. Without the presence check that became
// `ceph osd reweight <id> 0` and pushed all data off the OSD (#1772 review).
func TestBindOsdReweightTellsAbsentFromExplicitZero(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantPresent  bool
		wantReweight float64
	}{
		{name: "absent", body: `{}`, wantPresent: false, wantReweight: 0},
		{name: "null", body: `{"reweight":null}`, wantPresent: false, wantReweight: 0},
		{name: "explicit zero", body: `{"reweight":0}`, wantPresent: true, wantReweight: 0},
		{name: "0.95", body: `{"reweight":0.95}`, wantPresent: true, wantReweight: 0.95},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := nodes.OsdReqOpts{}

			present, err := bindOsdReweight(newPatchContext(test.body), &opts)

			require.NoError(t, err)
			require.Equal(t, test.wantPresent, present)
			require.Equal(t, test.wantReweight, opts.Reweight)
		})
	}
}

func TestBindOsdReweightRejectsMalformedJSON(t *testing.T) {
	_, err := bindOsdReweight(newPatchContext(`{"reweight":`), &nodes.OsdReqOpts{})
	require.Error(t, err)
}

// A body with no reweight must stop at parse time: no device lookup, no
// request to queue, and the handler answers "no update needed".
func TestParseUpdateOsdOptionsMarksAnAbsentReweight(t *testing.T) {
	c := newPatchContext(`{}`)
	c.Params = gin.Params{{Key: "nodeName", Value: "node1"}, {Key: "id", Value: "osd.2"}}
	h := &helper{c: c, handler: "updateNodeOsd"}

	require.NoError(t, h.parseUpdateOsdOptions())
	require.True(t, h.isReweightAbsent)
	require.Empty(t, h.osdReqOpts.Status.Desired)
}
