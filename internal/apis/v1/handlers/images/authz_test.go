package images

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
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

	return &helper{c: c, openstack: openstack.GetGlobalHelper(), handler: "test"}
}

func stubOidcRealmRoles(t *testing.T, roles []string, err error) {
	t.Helper()

	original := oidcRealmRoles
	oidcRealmRoles = func(string) (string, []string, error) {
		return "oidc-user", roles, err
	}
	t.Cleanup(func() { oidcRealmRoles = original })
}

func TestNodeTokenCallerIsAdmin(t *testing.T) {
	h := newTestHelper(t, auths.Oidc, map[string]string{
		"Node":          "n1",
		"Authorization": "Bearer " + nodes.GenToken("n1"),
	})

	require.NoError(t, h.resolveCaller())
	require.True(t, h.caller.isAdmin)
}

func TestOidcRealmAdminIsAdmin(t *testing.T) {
	stubOidcRealmRoles(t, []string{"admin"}, nil)
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})

	require.NoError(t, h.resolveCaller())
	require.True(t, h.caller.isAdmin)
}

func TestOidcUserWithoutRealmAdminHasNoProject(t *testing.T) {
	stubOidcRealmRoles(t, []string{"offline_access"}, nil)
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})

	require.NoError(t, h.resolveCaller())
	require.False(t, h.caller.isAdmin)
	require.False(t, h.caller.hasProject())
	require.False(t, h.caller.canWrite())
}

func TestUnverifiableOidcTokenIsForbidden(t *testing.T) {
	stubOidcRealmRoles(t, nil, errors.New("bad token"))
	h := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})

	require.True(t, isForbidden(h.resolveCaller()))
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
	stubOidcRealmRoles(t, []string{"offline_access"}, nil)
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

	stubOidcRealmRoles(t, []string{"offline_access"}, nil)
	unscoped := newTestHelper(t, auths.Oidc, map[string]string{"Authorization": "Bearer jwt"})
	require.NoError(t, unscoped.resolveCaller())
	require.Empty(t, unscoped.visibleProcessings(processings))
}
