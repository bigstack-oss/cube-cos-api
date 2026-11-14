package authz

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	conf "github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
)

const (
	realmAdminRole = "admin"
	realmAdminTtl  = 60 * time.Second
)

// realmRoleClient is the part of the Keycloak admin API used to look up a
// user's realm roles. The shared keycloak helper's client does not offer the
// role lookup and keeps its admin token on the shared helper, so this
// package keeps a client of its own.
type realmRoleClient interface {
	LoginAdmin(ctx context.Context, username, password, realm string) (*gocloak.JWT, error)
	GetUsers(ctx context.Context, token, realm string, params gocloak.GetUsersParams) ([]*gocloak.User, error)
	GetCompositeRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*gocloak.Role, error)
}

var (
	keycloakOnce   sync.Once
	keycloakClient realmRoleClient
)

// keycloak is a variable so tests can stand in for Keycloak.
var keycloak = func() realmRoleClient {
	keycloakOnce.Do(func() {
		opts := conf.Opts.Spec.Identity.Keycloak
		if opts.Ip == "" {
			opts.Ip = base.DataCenterVip
		}

		u := url.URL{
			Scheme: opts.Scheme,
			Host:   fmt.Sprintf("%s:%d", opts.Ip, opts.Port),
			Path:   opts.Path,
		}

		client := gocloak.NewClient(u.String())
		if opts.TlsInsecureSkipVerify {
			client.RestyClient().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		}

		keycloakClient = client
	})

	return keycloakClient
}

// lookupRealmAdmin asks Keycloak whether the user holds the realm admin role,
// either directly or through a group. A user Keycloak does not know is not an
// admin.
func lookupRealmAdmin(username string) (bool, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	defer cancel()

	client := keycloak()
	jwt, err := client.LoginAdmin(ctx, base.ServiceName, base.SystemSeed, auths.DefaultKeycloakRealm)
	if err != nil {
		return false, err
	}

	users, err := client.GetUsers(ctx, jwt.AccessToken, auths.DefaultKeycloakRealm, gocloak.GetUsersParams{
		Username: gocloak.StringP(username),
		Exact:    gocloak.BoolP(true),
	})
	if err != nil {
		return false, err
	}

	if len(users) == 0 || users[0].ID == nil {
		return false, nil
	}

	roles, err := client.GetCompositeRealmRolesByUserID(ctx, jwt.AccessToken, auths.DefaultKeycloakRealm, *users[0].ID)
	if err != nil {
		return false, err
	}

	return slices.ContainsFunc(roles, func(role *gocloak.Role) bool {
		return role != nil && gocloak.PString(role.Name) == realmAdminRole
	}), nil
}

type realmAdminEntry struct {
	isAdmin   bool
	expiresAt time.Time
}

// realmAdminCache keeps each user's answer for realmAdminTtl, so a burst of
// requests from one user costs one Keycloak lookup. Failed lookups are not
// kept.
type realmAdminCache struct {
	mu      sync.Mutex
	entries map[string]realmAdminEntry
	lookup  func(username string) (bool, error)
	now     func() time.Time
}

func newRealmAdminCache(lookup func(string) (bool, error)) *realmAdminCache {
	return &realmAdminCache{
		entries: map[string]realmAdminEntry{},
		lookup:  lookup,
		now:     time.Now,
	}
}

func (r *realmAdminCache) isAdmin(username string) (bool, error) {
	r.mu.Lock()
	entry, found := r.entries[username]
	r.mu.Unlock()
	if found && r.now().Before(entry.expiresAt) {
		return entry.isAdmin, nil
	}

	isAdmin, err := r.lookup(username)
	if err != nil {
		return false, err
	}

	r.mu.Lock()
	r.entries[username] = realmAdminEntry{isAdmin: isAdmin, expiresAt: r.now().Add(realmAdminTtl)}
	r.mu.Unlock()
	return isAdmin, nil
}

// IsRealmAdmin reports whether a Keycloak user holds the realm admin role.
// It is a variable so tests can stand in for Keycloak.
var IsRealmAdmin = newRealmAdminCache(lookupRealmAdmin).isAdmin
