package saml

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/auths"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	xrv "github.com/mattermost/xml-roundtrip-validator"
	dsig "github.com/russellhaering/goxmldsig"
	log "go-micro.dev/v5/logger"
)

/*
 * IdP-initiated SAML Single Logout (HTTP-POST binding).
 *
 * Keycloak's "api" client registers <root>/saml/slo as its POST logout URL and
 * runs front-channel logout through the browser. A LogoutRequest is honoured
 * only when it carries a valid IdP signature, names this SP's SLO URL as its
 * Destination, comes from the IdP in the metadata, is fresh, and has not been
 * seen before. crewjam/saml v0.5.1 can answer a LogoutRequest
 * (MakePostLogoutResponse) but has no SP-side validator for one, hence the
 * checks here, which mirror its own LogoutResponse validation.
 */

const (
	samlProtocolNS    = "urn:oasis:names:tc:SAML:2.0:protocol"
	sessionIndexClaim = "SessionIndex"

	// How long a session index named by a LogoutRequest is remembered on this
	// node: the lifetime of the cos_token cookie (crewjam's default session MaxAge).
	revokedSessionTTL = time.Hour
)

// The inline script of crewjam's LogoutResponse.Post form; the CSP below allows
// exactly this script and nothing else.
const logoutResponseScript = `document.getElementById('SAMLSubmitButton').style.visibility="hidden";` +
	`document.getElementById('SAMLResponseForm').submit();`

var (
	errSloBinding     = errors.New("only the HTTP-POST binding is supported")
	errSloNoRequest   = errors.New("no SAMLRequest in the form")
	errSloReplayed    = errors.New("LogoutRequest ID already seen")
	errSloUnsigned    = errors.New("LogoutRequest is not signed")
	errSloNotRequest  = errors.New("document is not a LogoutRequest")
	errSloNoSigCerts  = errors.New("no IdP signing certificate in the metadata")
	base64WhiteSpaces = regexp.MustCompile(`\s+`)

	// LogoutRequest IDs already processed, until they would be too old anyway.
	seenLogoutRequests = newExpirySet()
	// Session indexes (the cookie's SessionIndex claim) the IdP has logged out,
	// with when. Only cookies issued up to then are revoked: Keycloak can issue
	// the same SessionIndex again if its session outlives the logout (e.g. it
	// never got the LogoutResponse), and that new cookie is Keycloak's to judge.
	revokedSessions = newRevocationSet()
)

type logoutRequest struct {
	XMLName        xml.Name   `xml:"urn:oasis:names:tc:SAML:2.0:protocol LogoutRequest"`
	ID             string     `xml:"ID,attr"`
	Version        string     `xml:"Version,attr"`
	IssueInstant   time.Time  `xml:"IssueInstant,attr"`
	NotOnOrAfter   *time.Time `xml:"NotOnOrAfter,attr"`
	Destination    string     `xml:"Destination,attr"`
	Issuer         string     `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
	NameID         string     `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
	SessionIndexes []string   `xml:"urn:oasis:names:tc:SAML:2.0:protocol SessionIndex"`
}

// ServeSaml routes the unauthenticated /saml/* endpoints: the SLO URL to the
// logout handler, everything else to the ACS handler as before.
func ServeSaml() gin.HandlerFunc {
	acs := ServeAcs()
	slo := ServeSlo()
	return func(c *gin.Context) {
		if isSloPath(c.Request.URL.Path) {
			slo(c)
			return
		}

		acs(c)
	}
}

func isSloPath(path string) bool {
	return path == SpAuth.ServiceProvider.SloURL.Path
}

// ServeSlo handles an IdP-initiated LogoutRequest: it revokes the session
// indexes the request names, deletes the cos_token cookie when it belongs to
// one of them, and answers the IdP with a signed LogoutResponse (auto-POST form).
func ServeSlo() gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := readLogoutRequest(c.Request, &SpAuth.ServiceProvider, saml.TimeNow())
		if err != nil {
			log.Warnf("auth: rejected SAML logout request(%s)", samlFailureReason(err))
			bodies.SetBadRequest(c, errors.New("invalid SAML logout request"), nil)
			c.Abort()
			return
		}

		revokedAt := saml.TimeNow()
		for _, index := range req.SessionIndexes {
			revokedSessions.revoke(index, revokedAt)
		}

		if isCookieSessionNamedBy(c.Request, req) {
			err = SpAuth.Session.DeleteSession(c.Writer, c.Request)
			if err != nil {
				log.Errorf("auth: failed to delete the session named by a SAML logout request(%v)", err)
			}
		}
		log.Infof("auth: processed SAML logout request for %d session index(es)", len(req.SessionIndexes))

		writeLogoutResponse(c, req.ID, c.Request.PostForm.Get("RelayState"))
	}
}

func writeLogoutResponse(c *gin.Context, requestID, relayState string) {
	if SpAuth.ServiceProvider.GetSLOBindingLocation(saml.HTTPPostBinding) == "" {
		log.Errorf("auth: IdP metadata has no HTTP-POST SingleLogoutService, cannot answer the logout request")
		bodies.SetRedirect(c, auths.RedirectPath)
		c.Abort()
		return
	}

	form, err := SpAuth.ServiceProvider.MakePostLogoutResponse(requestID, relayState)
	if err != nil {
		log.Errorf("auth: failed to make SAML logout response(%v)", err)
		bodies.SetInternalServerError(c, errors.New("failed to make SAML logout response"))
		c.Abort()
		return
	}

	sum := sha256.Sum256([]byte(logoutResponseScript))
	c.Header("Content-Security-Policy", "default-src; script-src 'sha256-"+
		base64.StdEncoding.EncodeToString(sum[:])+"'; reflected-xss block; referrer no-referrer;")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/html; charset=utf-8", form)
	c.Abort()
}

func readLogoutRequest(r *http.Request, sp *saml.ServiceProvider, now time.Time) (*logoutRequest, error) {
	if r.Method != http.MethodPost {
		return nil, errSloBinding
	}

	err := r.ParseForm()
	if err != nil {
		return nil, fmt.Errorf("unable to parse form: %w", err)
	}

	encoded := r.PostForm.Get("SAMLRequest")
	if encoded == "" {
		return nil, errSloNoRequest
	}

	req, err := parseLogoutRequest(sp, encoded, now)
	if err != nil {
		return nil, err
	}

	if !seenLogoutRequests.addIfAbsent(req.ID, req.IssueInstant.Add(saml.MaxIssueDelay+saml.MaxClockSkew), now) {
		return nil, errSloReplayed
	}

	return req, nil
}

// parseLogoutRequest decodes a POST-binding SAMLRequest, verifies the IdP's
// enveloped signature and reads the request from the signed element only.
func parseLogoutRequest(sp *saml.ServiceProvider, encoded string, now time.Time) (*logoutRequest, error) {
	raw, err := base64.StdEncoding.DecodeString(base64WhiteSpaces.ReplaceAllString(encoded, ""))
	if err != nil {
		return nil, fmt.Errorf("unable to parse base64: %w", err)
	}

	err = xrv.Validate(strings.NewReader(string(raw)))
	if err != nil {
		return nil, errors.New("request failed XML round-trip validation (details withheld)")
	}

	doc := etree.NewDocument()
	err = doc.ReadFromBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("unable to parse XML: %w", err)
	}

	root := doc.Root()
	if root == nil || root.Tag != "LogoutRequest" || root.NamespaceURI() != samlProtocolNS {
		return nil, errSloNotRequest
	}

	signed, err := verifyIdpSignature(sp, root)
	if err != nil {
		return nil, err
	}

	signedDoc := etree.NewDocument()
	signedDoc.SetRoot(signed)
	signedXml, err := signedDoc.WriteToBytes()
	if err != nil {
		return nil, err
	}

	req := &logoutRequest{}
	err = xml.Unmarshal(signedXml, req)
	if err != nil {
		return nil, fmt.Errorf("unable to read LogoutRequest: %w", err)
	}

	err = validateLogoutRequest(sp, req, now)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func validateLogoutRequest(sp *saml.ServiceProvider, req *logoutRequest, now time.Time) error {
	if req.ID == "" {
		return errors.New("LogoutRequest has no ID")
	}
	if req.Version != "2.0" {
		return fmt.Errorf("unsupported LogoutRequest version %q", req.Version)
	}
	if req.Destination != sp.SloURL.String() {
		return fmt.Errorf("`Destination` does not match SloURL (expected %q)", sp.SloURL.String())
	}
	if sp.IDPMetadata == nil || strings.TrimSpace(req.Issuer) != sp.IDPMetadata.EntityID {
		return errors.New("`Issuer` is not the IdP in the metadata")
	}
	if req.IssueInstant.Add(saml.MaxIssueDelay).Before(now) {
		return fmt.Errorf("issueInstant expired at %s", req.IssueInstant.Add(saml.MaxIssueDelay))
	}
	if req.IssueInstant.After(now.Add(saml.MaxClockSkew)) {
		return errors.New("issueInstant is in the future")
	}
	if req.NotOnOrAfter != nil && !now.Before(req.NotOnOrAfter.Add(saml.MaxClockSkew)) {
		return fmt.Errorf("request expired at %s", req.NotOnOrAfter)
	}
	if len(req.SessionIndexes) == 0 && req.NameID == "" {
		return errors.New("LogoutRequest names no session")
	}

	return nil
}

// verifyIdpSignature checks the enveloped signature of el against each IdP
// signing certificate in the metadata, and returns the signed element.
func verifyIdpSignature(sp *saml.ServiceProvider, el *etree.Element) (*etree.Element, error) {
	if el.FindElement("./Signature") == nil {
		return nil, errSloUnsigned
	}

	err := rejectSha1(el)
	if err != nil {
		return nil, err
	}

	certs, err := idpSigningCerts(sp.IDPMetadata)
	if err != nil {
		return nil, err
	}

	// As crewjam does: a KeyInfo without a certificate (e.g. only a KeyName) is
	// dropped so the signature is checked against the metadata's certificates.
	el = el.Copy()
	if el.FindElement("./Signature/KeyInfo/X509Data/X509Certificate") == nil {
		sig := el.FindElement("./Signature")
		if keyInfo := sig.FindElement("KeyInfo"); keyInfo != nil {
			sig.RemoveChild(keyInfo)
		}
	}

	var lastErr error
	for _, cert := range certs {
		ctx := dsig.NewDefaultValidationContext(&dsig.MemoryX509CertificateStore{
			Roots: []*x509.Certificate{cert},
		})
		ctx.IdAttribute = "ID"
		if saml.Clock != nil {
			ctx.Clock = saml.Clock
		}

		signed, err := ctx.Validate(el)
		if err == nil {
			return signed, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("cannot validate signature on LogoutRequest: %w", lastErr)
}

// rejectSha1 refuses SHA-1 signature and digest methods. Keycloak signs this
// client's documents with RSA_SHA256.
func rejectSha1(el *etree.Element) error {
	for _, sig := range el.SelectElements("Signature") {
		for _, method := range append(
			sig.FindElements("./SignedInfo/SignatureMethod"),
			sig.FindElements("./SignedInfo/Reference/DigestMethod")...,
		) {
			alg := method.SelectAttrValue("Algorithm", "")
			if strings.HasSuffix(alg, "sha1") {
				return fmt.Errorf("SHA-1 signature or digest method %q is not accepted", alg)
			}
		}
	}

	return nil
}

func idpSigningCerts(md *saml.EntityDescriptor) ([]*x509.Certificate, error) {
	if md == nil {
		return nil, errSloNoSigCerts
	}

	certs := []*x509.Certificate{}
	for _, idp := range md.IDPSSODescriptors {
		for _, kd := range idp.KeyDescriptors {
			if kd.Use != "" && kd.Use != "signing" {
				continue
			}

			for _, c := range kd.KeyInfo.X509Data.X509Certificates {
				der, err := base64.StdEncoding.DecodeString(base64WhiteSpaces.ReplaceAllString(c.Data, ""))
				if err != nil {
					continue
				}
				cert, err := x509.ParseCertificate(der)
				if err != nil {
					continue
				}
				certs = append(certs, cert)
			}
		}
	}

	if len(certs) == 0 {
		return nil, errSloNoSigCerts
	}

	return certs, nil
}

// isCookieSessionNamedBy reports whether the request's cos_token session is one
// the LogoutRequest logs out: by SessionIndex, or by NameID when the request
// names no SessionIndex (SAML core 3.7.3.2).
func isCookieSessionNamedBy(r *http.Request, req *logoutRequest) bool {
	session, err := SpAuth.Session.GetSession(r)
	if err != nil || session == nil {
		return false
	}

	claims, ok := session.(samlsp.JWTSessionClaims)
	if !ok {
		return false
	}

	if len(req.SessionIndexes) == 0 {
		return req.NameID != "" && claims.Subject == req.NameID
	}

	for _, index := range claims.Attributes[sessionIndexClaim] {
		for _, named := range req.SessionIndexes {
			if index == named {
				return true
			}
		}
	}

	return false
}

// expirySet is a small process-local set whose keys expire.
type expirySet struct {
	mu   sync.Mutex
	keys map[string]time.Time
}

func newExpirySet() *expirySet {
	return &expirySet{keys: map[string]time.Time{}}
}

func (s *expirySet) add(key string, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(saml.TimeNow())
	s.keys[key] = until
}

// addIfAbsent adds key unless it is already present and unexpired.
func (s *expirySet) addIfAbsent(key string, until, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked(now)
	if _, found := s.keys[key]; found {
		return false
	}

	s.keys[key] = until
	return true
}

func (s *expirySet) has(key string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	until, found := s.keys[key]
	return found && now.Before(until)
}

func (s *expirySet) pruneLocked(now time.Time) {
	for key, until := range s.keys {
		if !now.Before(until) {
			delete(s.keys, key)
		}
	}
}

// revocationSet remembers when each key was revoked, for revokedSessionTTL.
type revocationSet struct {
	mu   sync.Mutex
	keys map[string]time.Time
}

func newRevocationSet() *revocationSet {
	return &revocationSet{keys: map[string]time.Time{}}
}

func (s *revocationSet) revoke(key string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, revokedAt := range s.keys {
		if !at.Before(revokedAt.Add(revokedSessionTTL)) {
			delete(s.keys, k)
		}
	}
	s.keys[key] = at
}

// revokedAt returns when key was revoked, if that is within revokedSessionTTL of now.
func (s *revocationSet) revokedAt(key string, now time.Time) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	at, found := s.keys[key]
	if !found || !now.Before(at.Add(revokedSessionTTL)) {
		return time.Time{}, false
	}

	return at, true
}

// revokes reports whether a cookie issued at issuedAt (JWT iat, whole seconds)
// predates the revocation of key.
func (s *revocationSet) revokes(key string, issuedAt, now time.Time) bool {
	at, found := s.revokedAt(key, now)
	return found && issuedAt.Unix() <= at.Unix()
}
