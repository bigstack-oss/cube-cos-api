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

type Options struct {
	IdentityProvider Provider `json:"identityProvider" yaml:"identityProvider"`
	ServiceProvider  Provider `json:"serviceProvider" yaml:"serviceProvider"`
}

type Provider struct {
	Host         `json:"host" yaml:"host"`
	MetadataPath string `json:"metadataPath" yaml:"metadataPath"`
}

type Host struct {
	Scheme                string `json:"scheme" yaml:"scheme"`
	VirtualIp             string `json:"virtualIp" yaml:"virtualIp"`
	Ip                    string `json:"ip" yaml:"ip"`
	Port                  int    `json:"port" yaml:"port"`
	TlsInsecureSkipVerify bool   `json:"tlsInsecureSkipVerify" yaml:"tlsInsecureSkipVerify"`
	Auth                  `json:"auth" yaml:"auth"`
}

type Auth struct {
	Key  string `json:"key" yaml:"key"`
	Cert string `json:"cert" yaml:"cert"`
}

func NewGlobalAuth(opts Options) error {
	keyPair, err := genApiServerCertKeyPair(
		opts.ServiceProvider.Host.Auth.Cert,
		opts.ServiceProvider.Host.Auth.Key,
	)
	if err != nil {
		log.Errorf("failed to generate api server cert key pair: %v", err)
		return err
	}

	idpMetadata, err := genIdentityProviderMetadata(opts)
	if err != nil {
		log.Errorf("failed to generate idp metadata: %v", err)
		return err
	}

	spMetadataUrl := genServiceProviderMetadataUrl(opts)
	SpAuth, err = samlsp.New(samlsp.Options{
		EntityID:    spMetadataUrl.String(),
		URL:         genRootUrl(opts),
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

func genIdentityProviderMetadata(opts Options) (*saml.EntityDescriptor, error) {
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.IdentityProvider.Host.TlsInsecureSkipVerify},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return samlsp.FetchMetadata(
		ctx,
		http.DefaultClient,
		genIdentityProviderMetadataUrl(opts),
	)
}

func genIdentityProviderMetadataUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.IdentityProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", definition.ControllerVip, opts.IdentityProvider.Host.Port),
		Path:   opts.IdentityProvider.MetadataPath,
	}
}

func genServiceProviderMetadataUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.ServiceProvider.Scheme,
		Host:   fmt.Sprintf("%s:%d", definition.ControllerVip, opts.ServiceProvider.Host.Port),
		Path:   opts.ServiceProvider.MetadataPath,
	}
}

func genRootUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.ServiceProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", definition.ControllerVip, opts.ServiceProvider.Host.Port),
	}
}
