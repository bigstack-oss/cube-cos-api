package nodes

import (
	"encoding/json"
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

// The API spec makes reweight required. A body without it (absent, null or a
// misspelt key) must be rejected at parse time, before any device lookup or
// queueing, instead of binding as 0 and draining the OSD.
func TestParseReweightOptionsRejectAMissingReweight(t *testing.T) {
	bodies := []string{`{}`, `{"reweight":null}`, `{"reweigh":0.5}`}

	for _, body := range bodies {
		t.Run("osd "+body, func(t *testing.T) {
			c := newPatchContext(body)
			c.Params = gin.Params{{Key: "nodeName", Value: "node1"}, {Key: "id", Value: "osd.2"}}
			h := &helper{c: c, handler: "updateNodeOsd"}

			require.EqualError(t, h.parseUpdateOsdOptions(), "reweight should be provided")
			require.Empty(t, h.osdReqOpts.Status.Desired)
		})

		t.Run("device osds "+body, func(t *testing.T) {
			c := newPatchContext(body)
			c.Params = gin.Params{{Key: "nodeName", Value: "node1"}, {Key: "device", Value: "sdc"}}
			h := &helper{c: c, handler: "updateNodeDeviceOsds"}

			require.EqualError(t, h.parseUpdateDeviceOsdsOptions(), "reweight should be provided")
			require.Empty(t, h.osdReqOptses)
		})
	}
}

// initHelper turns the parse error into a 400 before the request reaches
// validation or the operator.
func TestUpdateNodeOsdAnswers400ForAMissingReweight(t *testing.T) {
	for _, body := range []string{`{}`, `{"reweight":null}`, `{"reweigh":0.5}`} {
		t.Run(body, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "nodeName", Value: "node1"}, {Key: "id", Value: "osd.2"}}

			updateNodeOsd(c)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.Contains(t, rec.Body.String(), "reweight should be provided")
		})
	}
}

// A PATCH for an OSD on another node is forwarded with the whole OsdReqOpts as
// the body. Reweight has no omitempty, so an explicit 0 must survive that
// re-serialization and still count as provided on the owning node.
func TestProxiedOsdReqKeepsAnExplicitZeroReweight(t *testing.T) {
	forwarded, err := json.Marshal(nodes.OsdReqOpts{OsdId: "osd.2", Reweight: 0})
	require.NoError(t, err)

	opts := nodes.OsdReqOpts{}
	present, err := bindOsdReweight(newPatchContext(string(forwarded)), &opts)

	require.NoError(t, err)
	require.True(t, present)
	require.Equal(t, 0.0, opts.Reweight)
}
