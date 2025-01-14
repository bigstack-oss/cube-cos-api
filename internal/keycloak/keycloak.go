package keycloak

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"time"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

var (
	SamlAuth *samlsp.Middleware
)

type Saml struct {
	IdentityProvider Provider
	ServiceProvider  Provider
}

type Provider struct {
	Host
	MetadataPath string
}

type Host struct {
	Scheme      string
	VirtualIp   string
	Ip          string
	Port        int
	InsecureTls bool
	Auth
}

type Auth struct {
	Key  string
	Cert string
}

func NewGlobalSamlAuth(saml Saml) error {
	keyPair, err := genApiServerCertKeyPair(
		saml.ServiceProvider.Host.Auth.Cert,
		saml.ServiceProvider.Host.Auth.Key,
	)
	if err != nil {
		return err
	}

	idpMetadata, err := genIdpMetadata(saml)
	if err != nil {
		return err
	}

	spMetadataUrl := genSpSamlMetadataUrl(saml)
	SamlAuth, err = samlsp.New(samlsp.Options{
		EntityID:    spMetadataUrl.String(),
		URL:         genRootUrl(saml),
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
		SignRequest: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func genApiServerCertKeyPair(serverCert, serverKey string) (*tls.Certificate, error) {
	keyPair, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, err
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, err
	}

	return &keyPair, nil
}

func genIdpMetadata(saml Saml) (*saml.EntityDescriptor, error) {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: saml.IdentityProvider.Host.InsecureTls},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return samlsp.FetchMetadata(
		ctx,
		http.DefaultClient,
		genIdpSamlMetadataUrl(saml),
	)
}

func genIdpSamlMetadataUrl(saml Saml) url.URL {
	return url.URL{
		Scheme: saml.IdentityProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:10443", saml.IdentityProvider.Host.VirtualIp),
		Path:   definition.DefaultIdpSamlMetadataPath,
	}
}

func genSpSamlMetadataUrl(saml Saml) url.URL {
	return url.URL{
		Scheme: saml.ServiceProvider.Scheme,
		Host:   fmt.Sprintf("%s:%d", saml.ServiceProvider.Host.VirtualIp, saml.ServiceProvider.Host.Port),
		Path:   saml.ServiceProvider.MetadataPath,
	}
}

func genRootUrl(saml Saml) url.URL {
	return url.URL{
		Scheme: saml.ServiceProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", saml.ServiceProvider.Host.VirtualIp, saml.ServiceProvider.Host.Port),
	}
}
