package oidc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/config"
	"github.com/coreos/go-oidc"
)

func VerifyToken(token string) error {
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
		return err
	}

	ctx, cancel = context.WithTimeout(wait.CtxSeconds(10))
	oidcConf := &oidc.Config{SkipClientIDCheck: true}
	defer cancel()
	verifier := provider.Verifier(oidcConf)
	_, err = verifier.Verify(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func genRealmUrl() string {
	return fmt.Sprintf(
		"%s/realms/%s",
		config.Opts.Spec.Identity.Keycloak.Host,
		config.Opts.Spec.Identity.Keycloak.Realm,
	)
}
