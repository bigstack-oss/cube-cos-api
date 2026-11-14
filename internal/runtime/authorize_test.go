package runtime

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/apis"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/authz"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// newSamlTestRouter is newAuthTestRouter with a signed-in SAML user, whose
// realm admin role is answered by isRealmAdmin instead of Keycloak.
func newSamlTestRouter(t *testing.T, username string, isRealmAdmin func(string) (bool, error)) *gin.Engine {
	t.Helper()

	router := newAuthTestRouter(t)
	samlAuthRequest = func(c *gin.Context) {
		session := samlsp.JWTSessionClaims{Attributes: samlsp.Attributes{"username": {username}}}
		c.Request = c.Request.WithContext(samlsp.ContextWithSession(c.Request.Context(), session))
		c.Next()
	}

	original := authz.IsRealmAdmin
	authz.IsRealmAdmin = isRealmAdmin
	t.Cleanup(func() { authz.IsRealmAdmin = original })

	return router
}

func realmAdmin(isAdmin bool) func(string) (bool, error) {
	return func(string) (bool, error) { return isAdmin, nil }
}

func nodeHeader() http.Header {
	header := http.Header{}
	header.Set("Node", "node1")
	header.Set("Authorization", "Bearer "+nodes.GenToken("node1"))
	return header
}

type registeredRoute struct {
	apis.Handler
	path string
}

func registeredRoutes(t *testing.T) []registeredRoute {
	t.Helper()

	routes := []registeredRoute{}
	for _, handlers := range registeredHandlers() {
		for _, h := range handlers {
			routes = append(routes, registeredRoute{h, fillRoute(getUrlParentPath(h)+h.Path, "dc1")})
		}
	}
	require.Greater(t, len(routes), 100, "too few routes were registered")

	return routes
}

// expectStatus serves every registered route and checks the status the
// caller gets for each access level. A level missing from want is skipped.
func expectStatus(t *testing.T, router *gin.Engine, header http.Header, want map[apis.Access]int) {
	t.Helper()

	checked := map[apis.Access]int{}
	for _, r := range registeredRoutes(t) {
		status, found := want[r.Access]
		if !found {
			continue
		}

		require.Equal(t, status, serve(router, r.Method, r.path, header), "%s %s", r.Method, r.path)
		checked[r.Access]++
	}

	for access := range want {
		require.NotZero(t, checked[access], "no route of access %d was checked", access)
	}
}

func TestEveryWriteRouteDeclaresAccess(t *testing.T) {
	for _, r := range registeredRoutes(t) {
		if r.Method == http.MethodGet {
			continue
		}

		require.NotEqual(t, apis.AccessDefault, r.Access, "%s %s declares no access", r.Method, r.path)
	}
}

func TestSamlUserWithoutRealmAdminIsRefused(t *testing.T) {
	router := newSamlTestRouter(t, "bob", realmAdmin(false))

	expectStatus(t, router, nil, map[apis.Access]int{
		apis.AccessAdmin:   http.StatusForbidden,
		apis.AccessNode:    http.StatusForbidden,
		apis.AccessHandler: http.StatusOK,
		apis.AccessDefault: http.StatusOK,
	})
}

func TestSamlRealmAdminCallsAdminRoutes(t *testing.T) {
	router := newSamlTestRouter(t, "alice", realmAdmin(true))

	expectStatus(t, router, nil, map[apis.Access]int{
		apis.AccessAdmin:   http.StatusOK,
		apis.AccessNode:    http.StatusForbidden,
		apis.AccessHandler: http.StatusOK,
		apis.AccessDefault: http.StatusOK,
	})
}

func TestSamlUserIsLookedUpByUsername(t *testing.T) {
	asked := []string{}
	router := newSamlTestRouter(t, "alice", func(username string) (bool, error) {
		asked = append(asked, username)
		return true, nil
	})

	path := fillRoute(dcRoute("/firmwares"), "dc1")
	require.Equal(t, http.StatusOK, serve(router, http.MethodPost, path, nil))
	require.Equal(t, []string{"alice"}, asked)
}

func TestNodeCallsAdminAndNodeRoutes(t *testing.T) {
	router := newAuthTestRouter(t)
	original := authz.IsRealmAdmin
	authz.IsRealmAdmin = func(string) (bool, error) { return false, errors.New("nodes are not looked up") }
	t.Cleanup(func() { authz.IsRealmAdmin = original })

	expectStatus(t, router, nodeHeader(), map[apis.Access]int{
		apis.AccessAdmin:   http.StatusOK,
		apis.AccessNode:    http.StatusOK,
		apis.AccessHandler: http.StatusOK,
		apis.AccessDefault: http.StatusOK,
	})
}

func TestKeycloakFailureMakesAdminRoutesUnavailable(t *testing.T) {
	router := newSamlTestRouter(t, "alice", func(string) (bool, error) {
		return false, errors.New("keycloak down")
	})

	expectStatus(t, router, nil, map[apis.Access]int{
		apis.AccessAdmin:   http.StatusServiceUnavailable,
		apis.AccessNode:    http.StatusForbidden,
		apis.AccessHandler: http.StatusOK,
		apis.AccessDefault: http.StatusOK,
	})
	expectStatus(t, router, nodeHeader(), map[apis.Access]int{
		apis.AccessAdmin: http.StatusOK,
	})
}

func TestGpuInstanceConsoleNeedsAdmin(t *testing.T) {
	router := newSamlTestRouter(t, "bob", realmAdmin(false))
	path := fillRoute(dcRoute("/nodes/:nodeName/gpuCards/instances/:instanceId/console"), "dc1")
	require.Equal(t, http.StatusForbidden, serve(router, http.MethodGet, path, nil))

	router = newSamlTestRouter(t, "alice", realmAdmin(true))
	require.Equal(t, http.StatusOK, serve(router, http.MethodGet, path, nil))
}

func authorizeContext(method, authType string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, "/api/v1/datacenters/dc1/x", nil)
	c.Set("authType", authType)
	return c, rec
}

func TestKeystoneCallerCannotCallNodeRoutes(t *testing.T) {
	c, rec := authorizeContext(http.MethodPatch, auths.Openstack)

	authorize(apis.Handler{Method: http.MethodPatch, Access: apis.AccessNode})(c)
	require.True(t, c.IsAborted())
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUndeclaredWriteIsRefused(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		c, rec := authorizeContext(method, auths.Openstack)
		c.Set("isNode", true)

		authorize(apis.Handler{Method: method})(c)
		require.True(t, c.IsAborted(), method)
		require.Equal(t, http.StatusForbidden, rec.Code, method)
	}
}

func TestUndeclaredReadIsServed(t *testing.T) {
	c, _ := authorizeContext(http.MethodGet, auths.Saml)

	authorize(apis.Handler{Method: http.MethodGet})(c)
	require.False(t, c.IsAborted())
}

func TestRefusedCallerIsNamed(t *testing.T) {
	c, _ := authorizeContext(http.MethodPost, auths.Saml)
	session := samlsp.JWTSessionClaims{Attributes: samlsp.Attributes{"username": {"bob"}}}
	c.Request = c.Request.WithContext(samlsp.ContextWithSession(c.Request.Context(), session))
	require.Equal(t, "bob", callerName(c))

	c, _ = authorizeContext(http.MethodPost, auths.Oidc)
	c.Set("authUser", "alice")
	require.Equal(t, "alice", callerName(c))
}

func TestForwardedHeaderAloneIsNotANodeCall(t *testing.T) {
	router := newSamlTestRouter(t, "bob", realmAdmin(false))
	header := http.Header{}
	header.Set(nodes.ForwardedHeader, "node1")
	header.Set("Node", "node1")

	path := fillRoute(dcRoute("/integrations/storages/tasks"), "dc1")
	require.Equal(t, http.StatusForbidden, serve(router, http.MethodPatch, path, header))
}
