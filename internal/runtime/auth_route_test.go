package runtime

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	grafanapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/grafana"
	opensearchapi "github.com/bigstack-oss/cube-cos-api/internal/apis/v1/handlers/opensearch"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var allMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// newAuthTestRouter builds the production middleware chain with every
// control-converged route registered. Handlers answer 200 and the SAML
// fallback answers 401, so the status tells whether a request got past auth.
func newAuthTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	origSaml, origOidc := samlAuthRequest, verifyOidcToken
	samlAuthRequest = func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
	verifyOidcToken = func(string) (string, error) { return "", errors.New("no keycloak in tests") }
	t.Cleanup(func() { samlAuthRequest, verifyOidcToken = origSaml, origOidc })

	router := newGinRouter()
	for _, handlers := range registeredHandlers() {
		setRoleHandlersToRouter(router, stubHandlers(handlers))
	}

	return router
}

func registeredHandlers() map[string][]apis.Handler {
	prepareApiHandlersByRole()
	return apis.GetRoleHandlers(nodes.RoleControlConverged)
}

func stubHandlers(handlers []apis.Handler) []apis.Handler {
	stubs := make([]apis.Handler, 0, len(handlers))
	for _, h := range handlers {
		h.Func = func(c *gin.Context) { c.String(http.StatusOK, "handler") }
		stubs = append(stubs, h)
	}

	return stubs
}

// fillRoute turns a route template into a concrete path.
func fillRoute(template, dataCenter string) string {
	segments := strings.Split(template, "/")
	for i, s := range segments {
		switch {
		case s == ":DataCenter":
			segments[i] = dataCenter
		case strings.HasPrefix(s, ":"), strings.HasPrefix(s, "*"):
			segments[i] = "p1"
		}
	}

	return strings.Join(segments, "/")
}

func serve(router *gin.Engine, method, path string, header http.Header) int {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range header {
		req.Header[k] = v
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Code
}

func isTokenRoute(r route) bool {
	return r == route{http.MethodPost, dcRoute("/tokens")}
}

func TestAuthFreeRoutesAreServedWithoutCredentials(t *testing.T) {
	router := newAuthTestRouter(t)

	for r := range authFreeRoutes {
		path := fillRoute(r.path, "dc1")
		require.Equal(t, http.StatusOK, serve(router, r.method, path, nil), "%s %s", r.method, path)
	}
}

func TestTokenRouteSkipsSaml(t *testing.T) {
	router := newAuthTestRouter(t)

	path := fillRoute(dcRoute("/tokens"), "dc1")
	require.Equal(t, http.StatusOK, serve(router, http.MethodPost, path, nil))
}

func TestInternalNodeTokenIsAccepted(t *testing.T) {
	router := newAuthTestRouter(t)

	header := http.Header{}
	header.Set("Node", "node1")
	header.Set("Authorization", "Bearer "+nodes.GenToken("node1"))

	path := fillRoute(dcRoute("/nodes"), "dc1")
	require.Equal(t, http.StatusOK, serve(router, http.MethodGet, path, header))
}

// Every registered route outside the auth-free list must require credentials,
// whatever value the datacenter segment carries.
func TestRoutesRequireCredentials(t *testing.T) {
	router := newAuthTestRouter(t)

	checked := 0
	for _, ri := range router.Routes() {
		r := route{ri.Method, ri.Path}
		if authFreeRoutes[r] || isTokenRoute(r) || !strings.HasPrefix(ri.Path, apis.V1) {
			continue
		}

		for _, dataCenter := range []string{"dc1", "token"} {
			path := fillRoute(ri.Path, dataCenter)
			require.Equal(t, http.StatusUnauthorized, serve(router, ri.Method, path, nil), "%s %s", ri.Method, path)
			checked++
		}
	}
	require.Greater(t, checked, 100, "too few routes were checked")
}

func TestNodeNamedLikeAnAuthFreeModuleRequiresCredentials(t *testing.T) {
	router := newAuthTestRouter(t)

	for _, name := range []string{"grafana", "opensearch"} {
		path := strings.Replace(fillRoute(dcRoute("/nodes/:nodeName"), "dc1"), "p1", name, 1)
		require.Equal(t, http.StatusUnauthorized, serve(router, http.MethodGet, path, nil), path)
	}
}

func TestAuthFreeAndTokenRoutesOnlyAllowTheirMethod(t *testing.T) {
	router := newAuthTestRouter(t)

	allowed := map[route]bool{{http.MethodPost, dcRoute("/tokens")}: true}
	for r := range authFreeRoutes {
		allowed[r] = true
	}

	for r := range allowed {
		for _, method := range allMethods {
			if method == r.method {
				continue
			}

			path := fillRoute(r.path, "dc1")
			require.Equal(t, http.StatusUnauthorized, serve(router, method, path, nil), "%s %s", method, path)
		}
	}
}

func TestUnknownRouteRequiresCredentials(t *testing.T) {
	router := newAuthTestRouter(t)

	path := fillRoute(dcRoute("/no-such-route"), "dc1")
	require.Equal(t, http.StatusUnauthorized, serve(router, http.MethodGet, path, nil))
}

func TestAuthFreeRoutesMatchRegisteredHandlers(t *testing.T) {
	router := newAuthTestRouter(t)

	registered := map[route]bool{}
	for _, ri := range router.Routes() {
		registered[route{ri.Method, ri.Path}] = true
	}

	for r := range authFreeRoutes {
		require.True(t, registered[r], "auth-free route is not registered: %s %s", r.method, r.path)
	}
	require.True(t, registered[route{http.MethodPost, dcRoute("/tokens")}], "token route is not registered")
}

func TestGrafanaRoutesRequireCredentials(t *testing.T) {
	requireCredentialsForHandlers(t, grafanapi.Handlers)
}

func TestOpenSearchRoutesRequireCredentials(t *testing.T) {
	requireCredentialsForHandlers(t, opensearchapi.Handlers)
}

func requireCredentialsForHandlers(t *testing.T, handlers []apis.Handler) {
	t.Helper()
	router := newAuthTestRouter(t)

	for _, h := range handlers {
		r := route{h.Method, getUrlParentPath(h) + h.Path}
		require.False(t, authFreeRoutes[r], "%s %s", r.method, r.path)

		path := fillRoute(r.path, "dc1")
		require.Equal(t, http.StatusUnauthorized, serve(router, r.method, path, nil), "%s %s", r.method, path)
	}
}
