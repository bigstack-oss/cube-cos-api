package saml

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	log "go-micro.dev/v5/logger"
)

/*
 * Keycloak session check for cos_token.
 *
 * cos_token is a self-contained JWT, so a Keycloak-side sign-out (admin console,
 * account console, another SP's logout) does not reach it by itself. Each
 * SessionIndex in the cookie is "<user session id>::<client id>" (Keycloak's
 * SamlSessionUtils). The API keeps, per client, the set of user session ids
 * Keycloak lists for that client, refreshed at most once per sessionCheckTTL,
 * and treats a cookie whose user session id is no longer listed as signed out.
 *
 * Revocation window: a session ended in Keycloak stops working on a node within
 * sessionCheckTTL (30 s) plus one Keycloak call; an IdP-initiated LogoutRequest
 * (/saml/slo) revokes the cookies issued until then at once on the node that
 * receives it.
 *
 * Idle timeout: COS UI does not go back to Keycloak after the login, so Keycloak
 * drops the user session at the realm's SSO Session Idle even while the user
 * keeps using COS UI. The session list cannot tell that from a sign-out, so the
 * check only runs until sessionIdleMargin before the idle timeout, counted from
 * the cookie's issue time; after that the cookie is honoured until it expires.
 * A sign-out after that point takes effect when the cookie expires.
 *
 * Keycloak unreachable: fail open. The cookie is then honoured until it expires,
 * as before this check existed, and a warning is logged; Keycloak is not retried
 * for sessionCheckRetry. Failing closed would sign every COS UI user out during
 * a Keycloak restart, with no way to sign back in until Keycloak returns.
 */

const (
	sessionCheckTTL     = 30 * time.Second
	sessionCheckRetry   = 30 * time.Second
	sessionCheckTimeout = 3 * time.Second
	// A cookie issued within this margin of a snapshot may postdate it (JWT iat
	// is in whole seconds, and on 3cc another node's clock issued it).
	sessionIssueMargin = 5 * time.Second
	// Keycloak's idle clock may start before the cookie's iat (the browser posts
	// the assertion after Keycloak touched the session), or on another node's clock.
	sessionIdleMargin = time.Minute
	// Keycloak's default SSO Session Idle, used when the realm leaves it unset.
	defaultSsoSessionIdle = 30 * time.Minute

	sessionPageSize = 100
	sessionMaxPages = 100
)

// sessionLister lists the Keycloak user session ids that hold a session of the
// client, and reads the realm's SSO Session Idle in seconds.
type sessionLister interface {
	ListClientSessionIds(ctx context.Context, idOfClient string) (map[string]struct{}, error)
	SsoSessionIdle(ctx context.Context) (*int, error)
}

type clientSessions struct {
	mu       sync.Mutex
	ids      map[string]struct{}
	idle     time.Duration
	takenAt  time.Time
	failedAt time.Time
}

type sessionChecker struct {
	lister sessionLister
	now    func() time.Time

	mu      sync.Mutex
	clients map[string]*clientSessions
}

var sessionLiveness = newSessionChecker(&keycloakSessionLister{})

func newSessionChecker(lister sessionLister) *sessionChecker {
	return &sessionChecker{
		lister:  lister,
		now:     saml.TimeNow,
		clients: map[string]*clientSessions{},
	}
}

// isSessionAlive reports whether the cookie session may still be used: none of
// its session indexes was logged out by a LogoutRequest, and Keycloak still has
// the user sessions behind them.
func isSessionAlive(session samlsp.Session) bool {
	claims, ok := session.(samlsp.JWTSessionClaims)
	if !ok {
		return true
	}

	issuedAt := time.Unix(claims.IssuedAt, 0)
	for _, index := range claims.Attributes[sessionIndexClaim] {
		if revokedSessions.revokes(index, issuedAt, saml.TimeNow()) {
			return false
		}
		if !sessionLiveness.isAlive(index, issuedAt) {
			return false
		}
	}

	return true
}

// splitSessionIndex returns the Keycloak user session id and client id of a SessionIndex.
func splitSessionIndex(index string) (string, string, bool) {
	fragments := strings.Split(index, "::")
	if len(fragments) != 2 || fragments[0] == "" || fragments[1] == "" {
		return "", "", false
	}

	return fragments[0], fragments[1], true
}

func (s *sessionChecker) client(id string) *clientSessions {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, found := s.clients[id]
	if !found {
		c = &clientSessions{}
		s.clients[id] = c
	}

	return c
}

func (s *sessionChecker) isAlive(index string, issuedAt time.Time) bool {
	sessionId, clientId, ok := splitSessionIndex(index)
	if !ok {
		// Not a Keycloak session index; nothing to check it against.
		return true
	}

	c := s.client(clientId)
	c.mu.Lock()
	defer c.mu.Unlock()

	now := s.now()
	_, listed := c.ids[sessionId]
	fresh := !c.takenAt.IsZero() && now.Sub(c.takenAt) < sessionCheckTTL
	if fresh && (listed || issuedBefore(issuedAt, c.takenAt)) {
		return listed || pastIdleCheck(issuedAt, c.idle, now)
	}

	if !c.failedAt.IsZero() && now.Sub(c.failedAt) < sessionCheckRetry {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), sessionCheckTimeout)
	defer cancel()
	ids, idle, err := s.refresh(ctx, clientId)
	if err != nil {
		c.failedAt = now
		log.Warnf("auth: cannot check Keycloak sessions, COS UI sessions are not checked for %s(%v)", sessionCheckRetry, err)
		return true
	}

	c.ids = ids
	c.idle = idle
	c.takenAt = now
	c.failedAt = time.Time{}
	_, listed = c.ids[sessionId]
	return listed || !issuedBefore(issuedAt, c.takenAt) || pastIdleCheck(issuedAt, c.idle, now)
}

func (s *sessionChecker) refresh(ctx context.Context, clientId string) (map[string]struct{}, time.Duration, error) {
	ids, err := s.lister.ListClientSessionIds(ctx, clientId)
	if err != nil {
		return nil, 0, err
	}

	seconds, err := s.lister.SsoSessionIdle(ctx)
	if err != nil {
		return nil, 0, err
	}
	if seconds == nil || *seconds <= 0 {
		return ids, defaultSsoSessionIdle, nil
	}

	return ids, time.Duration(*seconds) * time.Second, nil
}

// pastIdleCheck reports whether Keycloak may have dropped the session of a
// cookie issued at issuedAt as idle by now, so its absence is no sign-out.
func pastIdleCheck(issuedAt time.Time, idle time.Duration, now time.Time) bool {
	return !now.Before(issuedAt.Add(idle - sessionIdleMargin))
}

// issuedBefore reports whether a cookie issued at issuedAt certainly predates a
// snapshot taken at takenAt, so its Keycloak session had to be in it.
func issuedBefore(issuedAt, takenAt time.Time) bool {
	return !issuedAt.Add(sessionIssueMargin).After(takenAt)
}

type clientUserSessionGetter interface {
	GetClientUserSessions(ctx context.Context, token, realm, idOfClient string, params ...gocloak.GetClientUserSessionsParams) ([]*gocloak.UserSessionRepresentation, error)
}

type realmGetter interface {
	GetRealm(ctx context.Context, token, realm string) (*gocloak.RealmRepresentation, error)
}

// keycloakSessionLister lists sessions with the API's Keycloak account, keeping
// its own admin token so it does not race other users of the global helper's.
type keycloakSessionLister struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func (k *keycloakSessionLister) ListClientSessionIds(ctx context.Context, idOfClient string) (map[string]struct{}, error) {
	h := keycloak.GetGlobalHelper()
	if h == nil || h.Client == nil {
		return nil, errors.New("keycloak helper is not initialised")
	}

	getter, ok := h.Client.(clientUserSessionGetter)
	if !ok {
		return nil, errors.New("keycloak client cannot list client user sessions")
	}

	token, err := k.adminToken(ctx, h)
	if err != nil {
		return nil, err
	}

	ids := map[string]struct{}{}
	for page := 0; page < sessionMaxPages; page++ {
		first, max := page*sessionPageSize, sessionPageSize
		sessions, err := getter.GetClientUserSessions(
			ctx,
			token,
			auths.DefaultKeycloakRealm,
			idOfClient,
			gocloak.GetClientUserSessionsParams{First: &first, Max: &max},
		)
		if err != nil {
			k.dropToken()
			return nil, err
		}

		for _, session := range sessions {
			if session != nil && session.ID != nil {
				ids[*session.ID] = struct{}{}
			}
		}
		if len(sessions) < sessionPageSize {
			return ids, nil
		}
	}

	return nil, fmt.Errorf("more than %d sessions for client %s", sessionPageSize*sessionMaxPages, idOfClient)
}

func (k *keycloakSessionLister) SsoSessionIdle(ctx context.Context) (*int, error) {
	h := keycloak.GetGlobalHelper()
	if h == nil || h.Client == nil {
		return nil, errors.New("keycloak helper is not initialised")
	}

	getter, ok := h.Client.(realmGetter)
	if !ok {
		return nil, errors.New("keycloak client cannot get the realm")
	}

	token, err := k.adminToken(ctx, h)
	if err != nil {
		return nil, err
	}

	realm, err := getter.GetRealm(ctx, token, auths.DefaultKeycloakRealm)
	if err != nil {
		k.dropToken()
		return nil, err
	}

	return realm.SsoSessionIdleTimeout, nil
}

func (k *keycloakSessionLister) adminToken(ctx context.Context, h *keycloak.Helper) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.token != "" && time.Now().Before(k.expiresAt) {
		return k.token, nil
	}

	jwt, err := h.Client.LoginAdmin(ctx, h.Options.Username, h.Options.Password, h.Options.Realm)
	if err != nil {
		return "", fmt.Errorf("keycloak login failed: %w", err)
	}

	k.token = jwt.AccessToken
	k.expiresAt = time.Now().Add(time.Duration(jwt.ExpiresIn)*time.Second - 10*time.Second)
	return k.token, nil
}

func (k *keycloakSessionLister) dropToken() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.token = ""
}
