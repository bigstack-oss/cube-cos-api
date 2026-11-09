package oidc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/coreos/go-oidc"
	log "go-micro.dev/v5/logger"
)

type claims struct {
	PreferredUsername string      `json:"preferred_username"`
	RealmAccess       realmAccess `json:"realm_access"`
}

type realmAccess struct {
	Roles []string `json:"roles"`
}

func VerifyToken(token string) (*claims, error) {
	provider, cancel := newProvider()
	defer cancel()
	if provider == nil {
		return nil, fmt.Errorf("oidc: failed to create oidc provider")
	}

	oidcToken, cancel := newToken(provider, token)
	defer cancel()
	if oidcToken == nil {
		return nil, fmt.Errorf("oidc: failed to create oidc token")
	}

	c := &claims{}
	err := oidcToken.Claims(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func newProvider() (*oidc.Provider, context.CancelFunc) {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	client := oidc.ClientContext(ctx, http.DefaultClient)
	provider, err := oidc.NewProvider(client, genRealmUrl())
	if err == nil {
		return provider, cancel
	}

	log.Errorf("oidc: failed to create oidc provider(%v)", err)
	return nil, cancel
}

func newToken(provider *oidc.Provider, token string) (*oidc.IDToken, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	conf := &oidc.Config{SkipClientIDCheck: true}
	verifier := provider.Verifier(conf)
	oidcToken, err := verifier.Verify(ctx, token)
	if err == nil {
		return oidcToken, cancel
	}

	return nil, cancel
}

func genRealmUrl() string {
	keycloak := &config.Opts.Spec.Identity.Keycloak
	// go-oidc requires the issuer in the discovery document to match this URL, and Keycloak
	// issues on the public host (KC_HOSTNAME, config_keycloak.cpp). The node resolves an FQDN to
	// the VIP itself (cube_scan's /etc/hosts entry), so this still reaches the local cluster.
	host := keycloak.Ip
	if host == "" {
		host = base.DataCenterPublicHost
	}

	u := url.URL{}
	u.Scheme = keycloak.Scheme
	u.Host = fmt.Sprintf("%s:%d", host, keycloak.Port)
	u.Path = fmt.Sprintf("%s/realms/%s", keycloak.Path, keycloak.Realm)
	return u.String()
}
