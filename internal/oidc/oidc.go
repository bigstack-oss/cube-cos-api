package oidc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/coreos/go-oidc"
)

type claims struct {
	PreferredUsername string `json:"preferred_username"`
}

func VerifyToken(token string) (*claims, error) {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx, cancel := context.WithTimeout(wait.CtxSeconds(10))
	client := oidc.ClientContext(ctx, http.DefaultClient)
	defer cancel()
	provider, err := oidc.NewProvider(client, genRealmUrl())
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(wait.CtxSeconds(10))
	oidcConf := &oidc.Config{SkipClientIDCheck: true}
	defer cancel()
	verifier := provider.Verifier(oidcConf)
	oidcToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	c := &claims{}
	err = oidcToken.Claims(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func genRealmUrl() string {
	ip := ""
	if config.Opts.Spec.Identity.Keycloak.Ip != "" {
		ip = config.Opts.Spec.Identity.Keycloak.Ip
	} else {
		ip = definition.DataCenterVip
	}

	return fmt.Sprintf(
		"%s://%s:%d%s/realms/%s",
		config.Opts.Spec.Identity.Keycloak.Scheme,
		ip,
		config.Opts.Spec.Identity.Keycloak.Port,
		config.Opts.Spec.Identity.Keycloak.Path,
		config.Opts.Spec.Identity.Keycloak.Realm,
	)
}
