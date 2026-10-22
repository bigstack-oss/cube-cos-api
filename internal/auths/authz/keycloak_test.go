package authz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/stretchr/testify/require"
)

type fakeKeycloak struct {
	loginErr  error
	users     []*gocloak.User
	usersErr  error
	roles     []*gocloak.Role
	rolesErr  error
	loginArgs []string
	params    gocloak.GetUsersParams
	realms    []string
	roleCalls int
}

func (f *fakeKeycloak) LoginAdmin(_ context.Context, username, password, realm string) (*gocloak.JWT, error) {
	f.loginArgs = []string{username, password, realm}
	if f.loginErr != nil {
		return nil, f.loginErr
	}

	return &gocloak.JWT{AccessToken: "admin-token"}, nil
}

func (f *fakeKeycloak) GetUsers(_ context.Context, token, realm string, params gocloak.GetUsersParams) ([]*gocloak.User, error) {
	f.params = params
	f.realms = append(f.realms, realm)
	return f.users, f.usersErr
}

func (f *fakeKeycloak) GetCompositeRealmRolesByUserID(_ context.Context, token, realm, userID string) ([]*gocloak.Role, error) {
	f.roleCalls++
	f.realms = append(f.realms, realm)
	return f.roles, f.rolesErr
}

func stubKeycloak(t *testing.T, fake *fakeKeycloak) {
	t.Helper()

	original := keycloak
	keycloak = func() realmRoleClient { return fake }
	t.Cleanup(func() { keycloak = original })
}

func keycloakUser(id string) []*gocloak.User {
	return []*gocloak.User{{ID: gocloak.StringP(id)}}
}

func realmRoles(names ...string) []*gocloak.Role {
	roles := []*gocloak.Role{}
	for _, name := range names {
		roles = append(roles, &gocloak.Role{Name: gocloak.StringP(name)})
	}

	return roles
}

func TestLookupRealmAdminFindsCompositeAdminRole(t *testing.T) {
	conf.Opts.Spec.Identity.Keycloak.Realm = "login-realm"
	fake := &fakeKeycloak{users: keycloakUser("u1"), roles: realmRoles("offline_access", "admin")}
	stubKeycloak(t, fake)

	isAdmin, err := lookupRealmAdmin("alice")
	require.NoError(t, err)
	require.True(t, isAdmin)
	require.Equal(t, []string{base.ServiceName, base.SystemSeed, "login-realm"}, fake.loginArgs)
	require.Equal(t, "alice", gocloak.PString(fake.params.Username))
	require.True(t, gocloak.PBool(fake.params.Exact))
	require.Equal(t, []string{auths.DefaultKeycloakRealm, auths.DefaultKeycloakRealm}, fake.realms)
}

func TestLookupRealmAdminWithoutAdminRole(t *testing.T) {
	stubKeycloak(t, &fakeKeycloak{users: keycloakUser("u1"), roles: realmRoles("offline_access", "uma_authorization")})

	isAdmin, err := lookupRealmAdmin("bob")
	require.NoError(t, err)
	require.False(t, isAdmin)
}

func TestLookupRealmAdminOfUnknownUserIsNotAdmin(t *testing.T) {
	fake := &fakeKeycloak{users: []*gocloak.User{}}
	stubKeycloak(t, fake)

	isAdmin, err := lookupRealmAdmin("ghost")
	require.NoError(t, err)
	require.False(t, isAdmin)
	require.Zero(t, fake.roleCalls)
}

func TestLookupRealmAdminFailsWhenKeycloakFails(t *testing.T) {
	cases := map[string]*fakeKeycloak{
		"login":       {loginErr: errors.New("down")},
		"users":       {usersErr: errors.New("down")},
		"role lookup": {users: keycloakUser("u1"), rolesErr: errors.New("down")},
	}

	for name, fake := range cases {
		t.Run(name, func(t *testing.T) {
			stubKeycloak(t, fake)

			_, err := lookupRealmAdmin("alice")
			require.Error(t, err)
		})
	}
}

type countingLookup struct {
	calls   int
	isAdmin bool
	err     error
}

func (l *countingLookup) lookup(string) (bool, error) {
	l.calls++
	return l.isAdmin, l.err
}

func TestRealmAdminCacheKeepsAnswerForTtl(t *testing.T) {
	lookup := &countingLookup{isAdmin: true}
	cache := newRealmAdminCache(lookup.lookup)
	now := time.Now()
	cache.now = func() time.Time { return now }

	for range 3 {
		isAdmin, err := cache.isAdmin("alice")
		require.NoError(t, err)
		require.True(t, isAdmin)
	}
	require.Equal(t, 1, lookup.calls)

	now = now.Add(realmAdminTtl - time.Second)
	_, err := cache.isAdmin("alice")
	require.NoError(t, err)
	require.Equal(t, 1, lookup.calls)

	lookup.isAdmin = false
	now = now.Add(time.Second)
	isAdmin, err := cache.isAdmin("alice")
	require.NoError(t, err)
	require.False(t, isAdmin)
	require.Equal(t, 2, lookup.calls)
}

func TestRealmAdminCacheIsPerUser(t *testing.T) {
	lookup := &countingLookup{isAdmin: true}
	cache := newRealmAdminCache(lookup.lookup)

	_, err := cache.isAdmin("alice")
	require.NoError(t, err)
	_, err = cache.isAdmin("bob")
	require.NoError(t, err)
	require.Equal(t, 2, lookup.calls)
}

func TestRealmAdminCacheDoesNotKeepFailures(t *testing.T) {
	lookup := &countingLookup{err: errors.New("down")}
	cache := newRealmAdminCache(lookup.lookup)

	_, err := cache.isAdmin("alice")
	require.Error(t, err)

	lookup.err = nil
	lookup.isAdmin = true
	isAdmin, err := cache.isAdmin("alice")
	require.NoError(t, err)
	require.True(t, isAdmin)
	require.Equal(t, 2, lookup.calls)
}
