package authz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/tokens"
	"github.com/stretchr/testify/require"
)

func newTestContext(t *testing.T, authType string) *gin.Context {
	t.Helper()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/settings", nil)
	if authType != "" {
		c.Set("authType", authType)
	}

	return c
}

func withSamlUser(c *gin.Context, username string) {
	session := samlsp.JWTSessionClaims{Attributes: samlsp.Attributes{"username": {username}}}
	c.Request = c.Request.WithContext(samlsp.ContextWithSession(c.Request.Context(), session))
}

type realmAdminStub struct {
	users   []string
	isAdmin bool
	err     error
}

func stubIsRealmAdmin(t *testing.T, isAdmin bool, err error) *realmAdminStub {
	t.Helper()

	stub := &realmAdminStub{isAdmin: isAdmin, err: err}
	original := IsRealmAdmin
	IsRealmAdmin = func(username string) (bool, error) {
		stub.users = append(stub.users, username)
		return stub.isAdmin, stub.err
	}
	t.Cleanup(func() { IsRealmAdmin = original })

	return stub
}

const keystoneToken = `{
	"token": {
		"user": {"id": "u1", "name": "alice"},
		"roles": [%s],
		"project": {"id": "p1", "name": "alpha", "domain": {"id": "d1", "name": "Default"}}
	}
}`

func stubOpenstackToken(t *testing.T, roles string) {
	t.Helper()

	var body any
	require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(keystoneToken, roles)), &body))

	original := openstackToken
	openstackToken = func(string) tokens.GetResult {
		result := tokens.GetResult{}
		result.Body = body
		return result
	}
	t.Cleanup(func() { openstackToken = original })
}

func TestNodeCallerIsAdmin(t *testing.T) {
	stub := stubIsRealmAdmin(t, false, nil)
	c := newTestContext(t, auths.Oidc)
	c.Request.Header.Set("Node", "n1")
	c.Set("isNode", true)

	caller, err := Resolve(c)
	require.NoError(t, err)
	require.True(t, caller.IsNode)
	require.True(t, caller.IsAdmin)
	require.Equal(t, "node:n1", caller.User)
	require.Empty(t, stub.users)
}

func TestOpenstackAdminRoleIsAdmin(t *testing.T) {
	for _, role := range []string{"admin", "service"} {
		t.Run(role, func(t *testing.T) {
			stubOpenstackToken(t, `{"id": "r1", "name": "`+role+`"}`)

			caller, err := Resolve(newTestContext(t, auths.Openstack))
			require.NoError(t, err)
			require.True(t, caller.IsAdmin)
		})
	}
}

func TestOpenstackMemberIsNotAdmin(t *testing.T) {
	stubOpenstackToken(t, `{"id": "r1", "name": "member"}, {"id": "r2", "name": "reader"}`)

	caller, err := Resolve(newTestContext(t, auths.Openstack))
	require.NoError(t, err)
	require.False(t, caller.IsAdmin)
	require.Equal(t, Caller{
		User:        "alice",
		Roles:       []string{"member", "reader"},
		ProjectId:   "p1",
		ProjectName: "alpha",
		DomainId:    "d1",
		DomainName:  "Default",
	}, caller)
}

func TestUnresolvableOpenstackTokenIsRefused(t *testing.T) {
	original := openstackToken
	openstackToken = func(string) tokens.GetResult {
		result := tokens.GetResult{}
		result.Err = errors.New("token not found")
		return result
	}
	t.Cleanup(func() { openstackToken = original })

	_, err := Resolve(newTestContext(t, auths.Openstack))
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrUnavailable)
}

func TestOidcCallerIsLookedUpInKeycloak(t *testing.T) {
	for _, isAdmin := range []bool{true, false} {
		stub := stubIsRealmAdmin(t, isAdmin, nil)
		c := newTestContext(t, auths.Oidc)
		c.Set("authUser", "alice")

		caller, err := Resolve(c)
		require.NoError(t, err)
		require.Equal(t, isAdmin, caller.IsAdmin)
		require.Equal(t, "alice", caller.User)
		require.Equal(t, []string{"alice"}, stub.users)
	}
}

func TestSamlCallerIsLookedUpInKeycloak(t *testing.T) {
	for _, isAdmin := range []bool{true, false} {
		stub := stubIsRealmAdmin(t, isAdmin, nil)
		c := newTestContext(t, auths.Saml)
		withSamlUser(c, "bob")

		caller, err := Resolve(c)
		require.NoError(t, err)
		require.Equal(t, isAdmin, caller.IsAdmin)
		require.Equal(t, "bob", caller.User)
		require.Equal(t, []string{"bob"}, stub.users)
	}
}

func TestKeycloakCallerWithoutUsernameIsNotAdmin(t *testing.T) {
	stub := stubIsRealmAdmin(t, true, nil)

	for _, authType := range []string{auths.Oidc, auths.Saml} {
		caller, err := Resolve(newTestContext(t, authType))
		require.NoError(t, err)
		require.False(t, caller.IsAdmin)
	}
	require.Empty(t, stub.users)
}

func TestKeycloakFailureIsUnavailable(t *testing.T) {
	stubIsRealmAdmin(t, false, errors.New("down"))
	c := newTestContext(t, auths.Saml)
	withSamlUser(c, "bob")

	_, err := Resolve(c)
	require.ErrorIs(t, err, ErrUnavailable)
}

func TestUnknownAuthTypeIsRefused(t *testing.T) {
	_, err := Resolve(newTestContext(t, ""))
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrUnavailable)
}

func TestResolveWorksCallerOutOncePerRequest(t *testing.T) {
	stub := stubIsRealmAdmin(t, true, nil)
	c := newTestContext(t, auths.Oidc)
	c.Set("authUser", "alice")

	for range 2 {
		caller, err := Resolve(c)
		require.NoError(t, err)
		require.True(t, caller.IsAdmin)
	}
	require.Len(t, stub.users, 1)
}

func TestResolveRetriesAfterFailure(t *testing.T) {
	stub := stubIsRealmAdmin(t, false, errors.New("down"))
	c := newTestContext(t, auths.Oidc)
	c.Set("authUser", "alice")

	_, err := Resolve(c)
	require.ErrorIs(t, err, ErrUnavailable)

	stub.err = nil
	stub.isAdmin = true
	caller, err := Resolve(c)
	require.NoError(t, err)
	require.True(t, caller.IsAdmin)
	require.Len(t, stub.users, 2)
}
