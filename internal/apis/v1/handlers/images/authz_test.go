package images

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/authz"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestHelper(t *testing.T, authType string, headers map[string]string) *helper {
	t.Helper()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/images", nil)
	for key, value := range headers {
		c.Request.Header.Set(key, value)
	}
	c.Set("authType", authType)
	if authType == auths.Oidc {
		c.Set("authUser", "oidc-user")
	}

	return &helper{c: c, openstack: openstack.GetGlobalHelper(), handler: "test"}
}

func stubRealmAdmin(t *testing.T, isAdmin bool, err error) {
	t.Helper()

	original := authz.IsRealmAdmin
	authz.IsRealmAdmin = func(string) (bool, error) {
		return isAdmin, err
	}
	t.Cleanup(func() { authz.IsRealmAdmin = original })
}

func TestNodeTokenCallerIsAdmin(t *testing.T) {
	h := newTestHelper(t, auths.Oidc, map[string]string{
		"Node":          "n1",
		"Authorization": "Bearer " + nodes.GenToken("n1"),
	})
	h.c.Set("isNode", true)

	require.NoError(t, h.resolveCaller())
	require.True(t, h.caller.isAdmin)
}

func TestOidcRealmAdminIsAdmin(t *testing.T) {
	stubRealmAdmin(t, true, nil)
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})

	require.NoError(t, h.resolveCaller())
	require.True(t, h.caller.isAdmin)
}

func TestOidcUserWithoutRealmAdminHasNoProject(t *testing.T) {
	stubRealmAdmin(t, false, nil)
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})

	require.NoError(t, h.resolveCaller())
	require.False(t, h.caller.isAdmin)
	require.False(t, h.caller.hasProject())
	require.False(t, h.caller.canWrite())
}

func TestKeycloakFailureLeavesCallerUnresolved(t *testing.T) {
	stubRealmAdmin(t, false, errors.New("keycloak down"))
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})

	err := h.resolveCaller()
	require.ErrorIs(t, err, authz.ErrUnavailable)
	require.False(t, isForbidden(err))
}

func TestUnresolvedCallerAnswersServiceUnavailable(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	setInitError(c, authz.ErrUnavailable)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestMaterialsOfMemberAreLimitedToOwnProject(t *testing.T) {
	h := newTestHelper(t, auths.Openstack, map[string]string{"X-Auth-Token": "alice-tok"})
	require.NoError(t, h.resolveCaller())

	projects, err := h.listProjects()
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "alpha", projects[0].Name)

	domains, err := h.listDomains()
	require.NoError(t, err)
	require.Equal(t, []string{"Default"}, domains)

	require.Empty(t, h.listReservedImages())
	require.Equal(t, []string{"private"}, h.listVisibilities())
}

func TestMaterialsOfCallerWithoutProjectAreEmpty(t *testing.T) {
	stubRealmAdmin(t, false, nil)
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})
	require.NoError(t, h.resolveCaller())

	projects, err := h.listProjects()
	require.NoError(t, err)
	require.Empty(t, projects)

	domains, err := h.listDomains()
	require.NoError(t, err)
	require.Empty(t, domains)
}

func TestProcessingsAreLimitedToOwnProject(t *testing.T) {
	processings := []images.Image{
		{Id: "own", Project: "alpha", Domain: "Default"},
		{Id: "other-project", Project: "beta", Domain: "Default"},
		{Id: "other-domain", Project: "alpha", Domain: "Other"},
	}

	member := newTestHelper(t, auths.Openstack, map[string]string{"X-Auth-Token": "alice-tok"})
	require.NoError(t, member.resolveCaller())
	visible := member.visibleProcessings(processings)
	require.Len(t, visible, 1)
	require.Equal(t, "own", visible[0].Id)

	admin := newTestHelper(t, auths.Openstack, map[string]string{"X-Auth-Token": "admin-tok"})
	require.NoError(t, admin.resolveCaller())
	require.Len(t, admin.visibleProcessings(processings), 3)

	stubRealmAdmin(t, false, nil)
	unscoped := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})
	require.NoError(t, unscoped.resolveCaller())
	require.Empty(t, unscoped.visibleProcessings(processings))
}
