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

	"github.com/Nerzal/gocloak/v13"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/keycloak"
	"github.com/bigstack-oss/cube-cos-api/internal/api"
	v1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
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
		Host:   fmt.Sprintf("%s:%d", v1.DataCenterVip, opts.IdentityProvider.Host.Port),
		Path:   opts.IdentityProvider.MetadataPath,
	}
}

func GetSamlClient(id string) (*gocloak.Client, error) {
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
	if err != nil {
		log.Errorf("runtime: failed to login admin for fetching saml client: %s", err.Error())
		return nil, err
	}

	clients, err := h.GetClients(
		v1.DefaultKeycloakRealm,
		gocloak.GetClientsParams{ClientID: gocloak.StringP(id)},
	)
	if err != nil {
		log.Errorf("runtime: failed to get clients: %s", err.Error())
		return nil, err
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("saml client not found")
	}

	return clients[0], nil
}

func CreateSamlMapper(id string, mapper gocloak.ProtocolMapperRepresentation) error {
	h := keycloak.GetGlobalHelper()
	err := h.LoginAdmin()
	if err != nil {
		log.Errorf("runtime: failed to login admin for mapper creation: %s", err.Error())
		return err
	}

	_, err = h.CreateClientProtocolMapper(v1.DefaultKeycloakRealm, id, mapper)
	if err == nil {
		return nil
	}
	if err.(*gocloak.APIError).Code == http.StatusConflict {
		return nil
	}

	return err
}

func GenServiceProviderMetadataUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.ServiceProvider.Scheme,
		Host:   fmt.Sprintf("%s:%d", v1.DataCenterVip, opts.ServiceProvider.Host.Port),
		Path:   opts.ServiceProvider.MetadataPath,
	}
}

func ServeAcs() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := c.Request.ParseForm()
		if err != nil {
			SpAuth.OnError(c.Writer, c.Request, err)
			return
		}

		err = checkTrackedRequest(c)
		if err != nil {
			SpAuth.OnError(c.Writer, c.Request, err)
			return
		}

		assertion, err := getAssertion(c)
		if err != nil {
			SpAuth.OnError(c.Writer, c.Request, err)
			return
		}

		err = createSession(c, assertion)
		if err != nil {
			SpAuth.OnError(c.Writer, c.Request, err)
			return
		}

		api.SetRedirect(
			c,
			SpAuth.ServiceProvider.DefaultRedirectURI,
		)
	}
}

func AuthRequest(c *gin.Context) {
	session, err := SpAuth.Session.GetSession(c.Request)
	if session != nil {
		c.Request = c.Request.WithContext(samlsp.ContextWithSession(c.Request.Context(), session))
		c.Next()
		return
	}
	if err != samlsp.ErrNoSession {
		SpAuth.OnError(c.Writer, c.Request, err)
		c.Abort()
		return
	}

	// for the request which isn't verified
	// if we try to redirect when the original request is the ACS URL
	// we'll end up in a loop.
	if isAcsPath(c.Request.URL.Path) {
		api.SetInternalServerError(c, errors.New("this path should not come here (SAML ACS)"))
		c.Abort()
		return
	}

	authReq, err := genAuthRequest()
	if err != nil {
		api.SetInternalServerError(c, err)
		c.Abort()
		return
	}

	relayState, err := genRelayState(c, authReq)
	if err != nil {
		api.SetInternalServerError(c, err)
		c.Abort()
		return
	}

	redirectURL, err := genRedirectUrl(authReq, relayState)
	if err != nil {
		api.SetInternalServerError(c, err)
		c.Abort()
		return
	}

	api.SetUnauthorized(c, errors.New(redirectURL.String()))
	c.Abort()
}

func genRootUrl(opts Options) url.URL {
	return url.URL{
		Scheme: opts.ServiceProvider.Host.Scheme,
		Host:   fmt.Sprintf("%s:%d", v1.DataCenterVip, opts.ServiceProvider.Host.Port),
	}
}

func checkTrackedRequest(c *gin.Context) error {
	trackedReqIndex := c.Request.Form.Get("RelayState")
	if trackedReqIndex == "" {
		return nil
	}

	_, err := SpAuth.RequestTracker.GetTrackedRequest(c.Request, trackedReqIndex)
	if err == nil {
		return SpAuth.RequestTracker.StopTrackingRequest(
			c.Writer,
			c.Request,
			trackedReqIndex,
		)
	}

	if err != http.ErrNoCookie || !SpAuth.ServiceProvider.AllowIDPInitiated {
		return err
	}

	return nil
}

func getAssertion(c *gin.Context) (*saml.Assertion, error) {
	reqIds := []string{}
	if SpAuth.ServiceProvider.AllowIDPInitiated {
		reqIds = append(reqIds, "")
	}

	for _, req := range SpAuth.RequestTracker.GetTrackedRequests(c.Request) {
		reqIds = append(reqIds, req.SAMLRequestID)
	}

	return SpAuth.ServiceProvider.ParseResponse(c.Request, reqIds)
}

func createSession(c *gin.Context, assertion *saml.Assertion) error {
	return SpAuth.Session.CreateSession(
		c.Writer,
		c.Request,
		assertion,
	)
}

func isAcsPath(path string) bool {
	return path == SpAuth.ServiceProvider.AcsURL.Path
}

func genAuthRequest() (*saml.AuthnRequest, error) {
	binding := saml.HTTPRedirectBinding
	bindingLocation := SpAuth.ServiceProvider.GetSSOBindingLocation(binding)
	return SpAuth.ServiceProvider.MakeAuthenticationRequest(
		bindingLocation,
		binding,
		SpAuth.ResponseBinding,
	)
}

func genRelayState(c *gin.Context, authReq *saml.AuthnRequest) (string, error) {
	return SpAuth.RequestTracker.TrackRequest(
		c.Writer,
		c.Request,
		authReq.ID,
	)
}

func genRedirectUrl(authReq *saml.AuthnRequest, relayState string) (*url.URL, error) {
	return authReq.Redirect(
		relayState,
		&SpAuth.ServiceProvider,
	)
}
