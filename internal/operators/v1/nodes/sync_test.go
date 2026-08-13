package node

import (
	oshttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/stretchr/testify/require"
)

func newPeer(t *testing.T, handler oshttp.HandlerFunc) nodes.Node {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	require.NoError(t, http.NewGlobalHelper())

	return nodes.Node{
		Hostname:   "cn07",
		DataCenter: "control",
		Protocol:   "http",
		Address:    strings.TrimPrefix(server.URL, "http://"),
	}
}

// A peer that cannot encode its own record answers 200 with a JSON content type
// and no body. The error must name that, instead of reporting the unmarshal
// failure that hides where the fault is.
func TestAskPeerNodeReportsAnEmptyBody(t *testing.T) {
	peer := newPeer(t, func(w oshttp.ResponseWriter, r *oshttp.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(oshttp.StatusOK)
	})

	o := &Operator{}
	node, err := o.askPeerNode(peer)

	require.Nil(t, node)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cn07")
	require.Contains(t, err.Error(), "empty body")
}

func TestAskPeerNodeReturnsThePeerRecord(t *testing.T) {
	peer := newPeer(t, func(w oshttp.ResponseWriter, r *oshttp.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(oshttp.StatusOK)
		_, _ = w.Write([]byte(`{"code":200,"status":"ok","data":{"hostname":"cn07","vcpu":{"totalCores":8}}}`))
	})

	o := &Operator{}
	node, err := o.askPeerNode(peer)

	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, "cn07", node.Hostname)
	require.Equal(t, float64(8), node.Vcpu.TotalCores)
}
