package v1

import (
	"fmt"

	"github.com/Nerzal/gocloak/v13"
)

const (
	Tokens = "tokens"
	Logout = "logout"
)

var (
	DefaultKeycloakRealm       = "master"
	DefaultIdpSamlMetadataPath = fmt.Sprintf("/auth/realms/%s/protocol/saml/descriptor", DefaultKeycloakRealm)
	DefaultSpSamlMetadataPath  = "/saml/metadata"
	DefaultApiServerKey        = "/var/www/certs/server.key"
	DefaultApiServerCert       = "/var/www/certs/server.cert"
	DefaultIdentifierFormat    = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"

	DefaultOidcClientId     = "token-connect"
	DefaultOidcClientSecret = ""
	DefaultOidcClientOpts   = gocloak.Client{
		ClientID:                  gocloak.StringP(DefaultOidcClientId),
		Protocol:                  gocloak.StringP("openid-connect"),
		PublicClient:              gocloak.BoolP(true),
		StandardFlowEnabled:       gocloak.BoolP(true),
		DirectAccessGrantsEnabled: gocloak.BoolP(true),
		Attributes: &map[string]string{
			"access.token.lifespan": "7200",
		},
	}

	LogoutRedirectUrl = ""
)
