package saml

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
	log "go-micro.dev/v5/logger"
)

var (
	SpAuth *samlsp.Middleware
)

type Spec struct {
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

func NewGlobalAuth(spec Spec) error {
	keyPair, err := genApiServerCertKeyPair(
		spec.ServiceProvider.Host.Auth.Cert,
		spec.ServiceProvider.Host.Auth.Key,
	)
	if err != nil {
		log.Errorf("failed to generate api server cert key pair: %v", err)
		return err
	}

	idpMetadata, err := genIdpMetadata(spec)
	if err != nil {
		log.Errorf("failed to generate idp metadata: %v", err)
		return err
	}

	spMetadataUrl := genSpMetadataUrl(spec)
	SpAuth, err = samlsp.New(samlsp.Options{
		EntityID:    spMetadataUrl.String(),
		URL:         genRootUrl(spec),
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
		SignRequest: true,
	})
	if err != nil {
		log.Errorf("failed to create saml auth: %v", err)
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

func genIdpMetadata(spec Spec) (*saml.EntityDescriptor, error) {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: spec.IdentityProvider.Host.InsecureTls},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return samlsp.FetchMetadata(
		ctx,
		http.DefaultClient,
		genIdpMetadataUrl(spec),
	)
}

func genIdpMetadataUrl(spec Spec) url.URL {
	return url.URL{
		Scheme: spec.IdentityProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", spec.IdentityProvider.Host.VirtualIp, spec.IdentityProvider.Host.Port),
		Path:   definition.DefaultIdpSamlMetadataPath,
	}
}

func genSpMetadataUrl(spec Spec) url.URL {
	return url.URL{
		Scheme: spec.ServiceProvider.Scheme,
		Host:   fmt.Sprintf("%s:%d", spec.ServiceProvider.Host.VirtualIp, spec.ServiceProvider.Host.Port),
		Path:   spec.ServiceProvider.MetadataPath,
	}
}

func genRootUrl(spec Spec) url.URL {
	return url.URL{
		Scheme: spec.ServiceProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", spec.ServiceProvider.Host.VirtualIp, spec.ServiceProvider.Host.Port),
	}
}
