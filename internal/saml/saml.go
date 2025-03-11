package saml

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/api"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
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

	idpMetadata, err := GenIdentityProviderMetadata(opts)
	if err != nil {
		log.Errorf("failed to generate idp metadata: %v", err)
		return err
	}

	spMetadataUrl := GenServiceProviderMetadataUrl(opts)
	SpAuth, err = samlsp.New(samlsp.Options{
		EntityID:           spMetadataUrl.String(),
		URL:                genRootUrl(opts),
		DefaultRedirectURI: "/",
		Key:                keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:        keyPair.Leaf,
		IDPMetadata:        idpMetadata,
		SignRequest:        true,
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

func GenIdentityProviderMetadata(opts Options) (*saml.EntityDescriptor, error) {
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
		GenIdentityProviderMetadataUrl(opts),
	)
}

func GenIdentityProviderMetadataUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.IdentityProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", definition.DataCenterVip, opts.IdentityProvider.Host.Port),
		Path:   opts.IdentityProvider.MetadataPath,
	}
}

func GenServiceProviderMetadataUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.ServiceProvider.Scheme,
		Host:   fmt.Sprintf("%s:%d", definition.DataCenterVip, opts.ServiceProvider.Host.Port),
		Path:   opts.ServiceProvider.MetadataPath,
	}
}

func genRootUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.ServiceProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", definition.DataCenterVip, opts.ServiceProvider.Host.Port),
	}
}

func ServeAcs() gin.HandlerFunc {
	return func(c *gin.Context) {
		m := SpAuth
		err := c.Request.ParseForm()
		if err != nil {
			m.OnError(c.Writer, c.Request, err)
			return
		}

		possibleRequestIDs := []string{}
		if m.ServiceProvider.AllowIDPInitiated {
			possibleRequestIDs = append(possibleRequestIDs, "")
		}

		trackedRequests := m.RequestTracker.GetTrackedRequests(c.Request)
		for _, tr := range trackedRequests {
			possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
		}

		assertion, err := m.ServiceProvider.ParseResponse(c.Request, possibleRequestIDs)
		if err != nil {
			m.OnError(c.Writer, c.Request, err)
			return
		}

		if trackedRequestIndex := c.Request.Form.Get("RelayState"); trackedRequestIndex != "" {
			_, err := m.RequestTracker.GetTrackedRequest(c.Request, trackedRequestIndex)
			if err != nil {
				if err != http.ErrNoCookie || !m.ServiceProvider.AllowIDPInitiated {
					m.OnError(c.Writer, c.Request, err)
					return
				}
			} else {
				if err := m.RequestTracker.StopTrackingRequest(c.Writer, c.Request, trackedRequestIndex); err != nil {
					m.OnError(c.Writer, c.Request, err)
					return
				}
			}
		}

		if err := m.Session.CreateSession(c.Writer, c.Request, assertion); err != nil {
			m.OnError(c.Writer, c.Request, err)
			return
		}

		api.SetRedirect(c, m.ServiceProvider.DefaultRedirectURI)
	}
}

func DoSamlAuth(c *gin.Context) {
	m := SpAuth
	session, err := m.Session.GetSession(c.Request)

	if session != nil {
		// verified
		c.Next()
		return
	}

	if err == samlsp.ErrNoSession {
		// not verified

		// If we try to redirect when the original request is the ACS URL we'll
		// end up in a loop.
		if c.Request.URL.Path == m.ServiceProvider.AcsURL.Path {
			api.SetInternalServerError(
				c,
				errors.New("this path should not come here (SAML ACS)"),
			)
			c.Abort()
			return
		}

		binding := saml.HTTPRedirectBinding
		bindingLocation := m.ServiceProvider.GetSSOBindingLocation(binding)
		authReq, err := m.ServiceProvider.MakeAuthenticationRequest(bindingLocation, binding, m.ResponseBinding)
		if err != nil {
			api.SetInternalServerError(c, err)
			c.Abort()
			return
		}

		relayState, err := m.RequestTracker.TrackRequest(c.Writer, c.Request, authReq.ID)
		if err != nil {
			api.SetInternalServerError(c, err)
			c.Abort()
			return
		}

		redirectURL, err := authReq.Redirect(relayState, &m.ServiceProvider)
		if err != nil {
			api.SetInternalServerError(c, err)
			c.Abort()
			return
		}
		api.SetUnauthorized(c, errors.New(redirectURL.String()))
		c.Abort()
		return
	}

	m.OnError(c.Writer, c.Request, err)
	c.Abort()
}
