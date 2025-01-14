package v1

import "fmt"

var (
	DefaultKeycloakRealm       = "master"
	DefaultIdpSamlMetadataPath = fmt.Sprintf("/auth/realms/%s/protocol/saml/descriptor", DefaultKeycloakRealm)
	DefaultSpSamlMetadataPath  = "/api/v1/saml/metadata"
	DefaultApiServerKey        = "/var/www/certs/server.key"
	DefaultApiServerCert       = "/var/www/certs/server.cert"
	DefaultIdentifierFormat    = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
)
