package supportfiles

import (
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/support"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Pins the process-wide local zone the way runtime.initSystemTime would, and
// restores it so the offset does not leak between tests.
func useLocalZone(t *testing.T, offsetSeconds int) {
	t.Helper()

	origSeconds := time.LocalZoneSeconds
	origZone := time.LocalFixedZone

	time.LocalZoneSeconds = offsetSeconds
	time.LocalFixedZone = ostime.FixedZone("", offsetSeconds)

	t.Cleanup(func() {
		time.LocalZoneSeconds = origSeconds
		time.LocalFixedZone = origZone
	})
}

// Asserts the value is RFC3339 carrying the node's offset — not a wall clock
// mislabeled with a fixed "+00:00" — and stamped at roughly now.
func requireStampedInLocalZone(t *testing.T, createdAt string, offsetSeconds int) {
	t.Helper()

	parsed, err := ostime.Parse(time.FormatRFC3339, createdAt)
	require.NoError(t, err, "createdAt must be RFC3339")

	_, offset := parsed.Zone()
	require.Equal(t, offsetSeconds, offset)
	require.WithinDuration(t, ostime.Now(), parsed, ostime.Minute)
}

func newTestContext(t *testing.T, body string) *gin.Context {
	t.Helper()

	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(nethttp.MethodPost, "/supportFiles", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	return c
}

func TestParseHostsStampsCreatedAtWithTheNodeLocalOffset(t *testing.T) {
	offsets := map[string]int{
		"east of UTC": 8 * 60 * 60,
		"west of UTC": -5 * 60 * 60,
		"at UTC":      0,
	}

	for name, offsetSeconds := range offsets {
		t.Run(name, func(t *testing.T) {
			useLocalZone(t, offsetSeconds)

			h := &helper{c: newTestContext(t, `{"hosts":["node-1"],"description":"probe"}`)}

			require.NoError(t, h.parseHosts())
			require.Equal(t, []string{"node-1"}, h.fileReq.Hosts)
			requireStampedInLocalZone(t, h.fileReq.CreatedAt, offsetSeconds)
		})
	}
}

// A createdAt supplied by the delegating control node is authoritative — every
// node in the set must file under one group, so it is not re-stamped locally.
func TestSetSupportFileReqKeepsACallerSuppliedCreatedAt(t *testing.T) {
	useLocalZone(t, 8*60*60)

	const delegated = "2025-07-24T15:02:37+08:00"
	h := &helper{fileReq: support.FileRequest{CreatedAt: delegated}}

	h.setSupportFileReq()

	require.Equal(t, delegated, h.fileReq.CreatedAt)
	require.Equal(t, delegated, h.file.Status.CreatedAt)
	require.Contains(t, h.file.Group, delegated)
}

func TestSetSupportFileReqStampsAMissingCreatedAtWithTheNodeLocalOffset(t *testing.T) {
	useLocalZone(t, -5*60*60)

	h := &helper{fileReq: support.FileRequest{Description: "probe"}}

	h.setSupportFileReq()

	requireStampedInLocalZone(t, h.fileReq.CreatedAt, -5*60*60)
	require.Equal(t, h.fileReq.CreatedAt, h.file.Status.CreatedAt)
	require.Equal(t, status.Creating, h.file.Status.Current)
}
