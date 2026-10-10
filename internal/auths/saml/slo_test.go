package saml

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/stretchr/testify/require"
)

const (
	testSpRoot    = "https://10.0.0.1:443"
	testSloUrl    = testSpRoot + "/saml/slo"
	testIdpEntity = "https://10.0.0.1:10443/auth/realms/master"
	testIdpSlo    = "https://10.0.0.1:10443/auth/realms/master/protocol/saml"
	testSession   = "user-session-1::client-1"
)

type testParty struct {
	key  *rsa.PrivateKey
	cert *x509.Certificate
}

func newTestParty(t *testing.T, name string) testParty {
	t.Helper()

	key := testKey(t)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return testParty{key: key, cert: cert}
}

func testIdpMetadata(t *testing.T, idp testParty) *saml.EntityDescriptor {
	t.Helper()

	md := fmt.Sprintf(`<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="%s">`+
		`<md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">`+
		`<md:KeyDescriptor use="signing"><ds:KeyInfo><ds:KeyName>k1</ds:KeyName><ds:X509Data><ds:X509Certificate>%s</ds:X509Certificate></ds:X509Data></ds:KeyInfo></md:KeyDescriptor>`+
		`<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="%s"/>`+
		`<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="%s"/>`+
		`</md:IDPSSODescriptor></md:EntityDescriptor>`,
		testIdpEntity, base64.StdEncoding.EncodeToString(idp.cert.Raw), testIdpSlo, testIdpSlo)
	entity := &saml.EntityDescriptor{}
	require.NoError(t, xml.Unmarshal([]byte(md), entity))
	return entity
}

// sloFixture installs a SpAuth trusting idp, and clean process-local state.
type sloFixture struct {
	sp  testParty
	idp testParty
}

func newSloFixture(t *testing.T) *sloFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	f := &sloFixture{sp: newTestParty(t, "sp"), idp: newTestParty(t, "idp")}
	root, err := url.Parse(testSpRoot)
	require.NoError(t, err)

	prevAuth, prevSeen, prevRevoked, prevLiveness := SpAuth, seenLogoutRequests, revokedSessions, sessionLiveness
	t.Cleanup(func() {
		SpAuth, seenLogoutRequests, revokedSessions, sessionLiveness = prevAuth, prevSeen, prevRevoked, prevLiveness
	})

	SpAuth, err = samlsp.New(samlsp.Options{
		CookieName:         "cos_token",
		EntityID:           testSpRoot + "/saml/metadata",
		URL:                *root,
		DefaultRedirectURI: "/",
		Key:                f.sp.key,
		Certificate:        f.sp.cert,
		IDPMetadata:        testIdpMetadata(t, f.idp),
		SignRequest:        true,
	})
	require.NoError(t, err)
	seenLogoutRequests = newExpirySet()
	revokedSessions = newRevocationSet()
	sessionLiveness = newSessionChecker(&fakeLister{ids: map[string]struct{}{"user-session-1": {}}})
	return f
}

// cookie returns a cos_token issued for an assertion with the given SessionIndex.
func (f *sloFixture) cookie(t *testing.T, sessionIndex string) *http.Cookie {
	t.Helper()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, testSpRoot+"/saml/acs", nil)
	err := SpAuth.Session.CreateSession(w, r, &saml.Assertion{
		Subject:         &saml.Subject{NameID: &saml.NameID{Value: "rt-user"}},
		AuthnStatements: []saml.AuthnStatement{{SessionIndex: sessionIndex}},
	})
	require.NoError(t, err)
	for _, c := range w.Result().Cookies() {
		if c.Name == "cos_token" {
			return c
		}
	}
	t.Fatal("no cos_token issued")
	return nil
}

type logoutReq struct {
	id           string
	destination  string
	issuer       string
	issueInstant time.Time
	sessionIndex string
	signer       *testParty
	keyName      bool // KeyInfo with a KeyName only, no certificate
	sha1         bool // RSA-SHA1 signature and SHA-1 digest
	tamper       bool // change the SessionIndex after signing
	root         string
}

func (f *sloFixture) logoutRequest(t *testing.T, mod func(*logoutReq)) string {
	t.Helper()

	req := logoutReq{
		id:           "ID_" + base64.RawURLEncoding.EncodeToString(randomBytes(t)),
		destination:  testSloUrl,
		issuer:       testIdpEntity,
		issueInstant: time.Now().UTC(),
		sessionIndex: testSession,
		signer:       &f.idp,
		root:         "LogoutRequest",
	}
	if mod != nil {
		mod(&req)
	}

	el := etree.NewElement("samlp:" + req.root)
	el.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	el.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	el.CreateAttr("ID", req.id)
	el.CreateAttr("Version", "2.0")
	el.CreateAttr("IssueInstant", req.issueInstant.Format(time.RFC3339Nano))
	el.CreateAttr("Destination", req.destination)
	el.CreateElement("saml:Issuer").SetText(req.issuer)
	el.CreateElement("saml:NameID").SetText("rt-user")
	el.CreateElement("samlp:SessionIndex").SetText(req.sessionIndex)

	if req.signer != nil {
		ctx, err := dsig.NewSigningContext(req.signer.key, [][]byte{req.signer.cert.Raw})
		require.NoError(t, err)
		ctx.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
		method := dsig.RSASHA256SignatureMethod
		if req.sha1 {
			method = dsig.RSASHA1SignatureMethod
		}
		require.NoError(t, ctx.SetSignatureMethod(method))
		el, err = ctx.SignEnveloped(el)
		require.NoError(t, err)

		// Keycloak places the Signature right after the Issuer.
		moved := etree.NewElement(el.Tag)
		moved.Space = el.Space
		moved.Attr = append(moved.Attr, el.Attr...)
		sig := el.FindElement("./Signature").Copy()
		for _, child := range el.ChildElements() {
			if child.Tag == "Signature" {
				continue
			}
			moved.AddChild(child.Copy())
			if child.Tag == "Issuer" {
				moved.AddChild(sig)
			}
		}
		el = moved
		if req.keyName {
			keyInfo := sig.FindElement("KeyInfo")
			sig.RemoveChild(keyInfo)
			sig.CreateElement("ds:KeyInfo").CreateElement("ds:KeyName").SetText("k1")
		}
	}
	if req.tamper {
		el.FindElement("./SessionIndex").SetText("user-session-2::client-1")
	}

	doc := etree.NewDocument()
	doc.SetRoot(el)
	raw, err := doc.WriteToBytes()
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(raw)
}

func isRevoked(index string) bool {
	_, found := revokedSessions.revokedAt(index, saml.TimeNow())
	return found
}

func randomBytes(t *testing.T) []byte {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}

func serveSaml(method, path string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
	router := gin.New()
	router.Any("/saml/*any", ServeSaml())

	var r *http.Request
	if method == http.MethodPost {
		r = httptest.NewRequest(method, testSpRoot+path, strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		r = httptest.NewRequest(method, testSpRoot+path+"?"+form.Encode(), nil)
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func deletedCosToken(w *httptest.ResponseRecorder) bool {
	for _, c := range w.Result().Cookies() {
		if c.Name == "cos_token" && c.Value == "" && c.Expires.Before(time.Unix(2, 0)) {
			return true
		}
	}
	return false
}

var formField = regexp.MustCompile(`name="(SAMLResponse|RelayState)" value="([^"]*)"`)

// logoutResponse reads the auto-POST form and checks the LogoutResponse in it.
func (f *sloFixture) logoutResponse(t *testing.T, w *httptest.ResponseRecorder, requestID string) string {
	t.Helper()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Contains(t, w.Header().Get("Content-Security-Policy"), "script-src 'sha256-")
	body := w.Body.String()
	require.Contains(t, body, `action="`+testIdpSlo+`"`)
	require.Contains(t, body, logoutResponseScript)

	// The CSP must allow exactly the script the form carries, or it never submits.
	script := regexp.MustCompile(`<script>(.*)</script>`).FindStringSubmatch(body)
	require.Len(t, script, 2)
	sum := sha256.Sum256([]byte(script[1]))
	require.Contains(t, w.Header().Get("Content-Security-Policy"),
		"script-src 'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")

	fields := map[string]string{}
	for _, m := range formField.FindAllStringSubmatch(body, -1) {
		fields[m[1]] = m[2]
	}
	raw, err := base64.StdEncoding.DecodeString(html.UnescapeString(fields["SAMLResponse"]))
	require.NoError(t, err)

	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(raw))
	require.Equal(t, "LogoutResponse", doc.Root().Tag)
	require.Equal(t, requestID, doc.Root().SelectAttrValue("InResponseTo", ""))
	require.Equal(t, testIdpSlo, doc.Root().SelectAttrValue("Destination", ""))
	require.Contains(t, string(raw), "urn:oasis:names:tc:SAML:2.0:status:Success")

	// Keycloak requires the SP's documents signed (saml.client.signature).
	ctx := dsig.NewDefaultValidationContext(&dsig.MemoryX509CertificateStore{Roots: []*x509.Certificate{f.sp.cert}})
	ctx.IdAttribute = "ID"
	_, err = ctx.Validate(doc.Root())
	require.NoError(t, err)
	return html.UnescapeString(fields["RelayState"])
}

func requestID(t *testing.T, encoded string) string {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(raw))
	return doc.Root().SelectAttrValue("ID", "")
}

func TestServeSloSignedLogoutRequest(t *testing.T) {
	for name, mod := range map[string]func(*logoutReq){
		"keycloak shaped":             nil,
		"keyinfo without cert":        func(r *logoutReq) { r.keyName = true },
		"slightly ahead of our clock": func(r *logoutReq) { r.issueInstant = time.Now().Add(time.Minute) },
	} {
		t.Run(name, func(t *testing.T) {
			f := newSloFixture(t)
			encoded := f.logoutRequest(t, mod)

			w := serveSaml(http.MethodPost, "/saml/slo", url.Values{
				"SAMLRequest": {encoded},
				"RelayState":  {"rs-1"},
			}, f.cookie(t, testSession))

			require.Equal(t, "rs-1", f.logoutResponse(t, w, requestID(t, encoded)))
			require.True(t, deletedCosToken(w), "cos_token must be deleted")
			require.True(t, isRevoked(testSession))
		})
	}
}

func TestServeSloRejects(t *testing.T) {
	other := newTestParty(t, "other")
	cases := map[string]struct {
		mod    func(*logoutReq)
		method string
		form   func(encoded string) url.Values
	}{
		"unsigned":              {mod: func(r *logoutReq) { r.signer = nil }},
		"signed by another key": {mod: func(r *logoutReq) { r.signer = &other }},
		"changed after signing": {mod: func(r *logoutReq) { r.tamper = true }},
		"wrong destination":     {mod: func(r *logoutReq) { r.destination = "https://10.0.0.2:443/saml/slo" }},
		"wrong issuer":          {mod: func(r *logoutReq) { r.issuer = "https://10.0.0.1:10443/auth/realms/other" }},
		"expired":               {mod: func(r *logoutReq) { r.issueInstant = time.Now().Add(-10 * time.Minute) }},
		"too far in the future": {mod: func(r *logoutReq) { r.issueInstant = time.Now().Add(10 * time.Minute) }},
		"not a logout request":  {mod: func(r *logoutReq) { r.root = "LogoutResponse" }},
		"redirect binding":      {method: http.MethodGet},
		"no SAMLRequest": {form: func(encoded string) url.Values {
			return url.Values{"SAMLResponse": {encoded}}
		}},
		"SHA-1 signature": {mod: func(r *logoutReq) { r.sha1 = true }},
		"not base64":      {form: func(string) url.Values { return url.Values{"SAMLRequest": {"%%%"}} }},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := newSloFixture(t)
			encoded := f.logoutRequest(t, tc.mod)
			form := url.Values{"SAMLRequest": {encoded}}
			if tc.form != nil {
				form = tc.form(encoded)
			}
			method := http.MethodPost
			if tc.method != "" {
				method = tc.method
			}

			w := serveSaml(method, "/saml/slo", form, f.cookie(t, testSession))

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.False(t, deletedCosToken(w), "cos_token must be kept")
			require.NotContains(t, w.Body.String(), "SAMLResponse")
			require.False(t, isRevoked(testSession))
		})
	}
}

func TestServeSloUnknownSession(t *testing.T) {
	f := newSloFixture(t)
	encoded := f.logoutRequest(t, func(r *logoutReq) { r.sessionIndex = "user-session-9::client-1" })

	w := serveSaml(http.MethodPost, "/saml/slo", url.Values{"SAMLRequest": {encoded}}, f.cookie(t, testSession))

	// The IdP is told the logout succeeded; another session's cookie is left alone.
	f.logoutResponse(t, w, requestID(t, encoded))
	require.False(t, deletedCosToken(w))
	require.True(t, isRevoked("user-session-9::client-1"))
	require.False(t, isRevoked(testSession))
}

func TestServeSloWithoutCookie(t *testing.T) {
	f := newSloFixture(t)
	encoded := f.logoutRequest(t, nil)

	w := serveSaml(http.MethodPost, "/saml/slo", url.Values{"SAMLRequest": {encoded}}, nil)

	f.logoutResponse(t, w, requestID(t, encoded))
	require.True(t, isRevoked(testSession))
}

func TestServeSloReplay(t *testing.T) {
	f := newSloFixture(t)
	encoded := f.logoutRequest(t, nil)
	form := url.Values{"SAMLRequest": {encoded}}

	first := serveSaml(http.MethodPost, "/saml/slo", form, f.cookie(t, testSession))
	f.logoutResponse(t, first, requestID(t, encoded))

	again := serveSaml(http.MethodPost, "/saml/slo", form, f.cookie(t, testSession))
	require.Equal(t, http.StatusBadRequest, again.Code)
	require.False(t, deletedCosToken(again))
}

// signedElement decodes a request built by logoutRequest back into its root element.
func signedElement(t *testing.T, encoded string) *etree.Element {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	doc := etree.NewDocument()
	require.NoError(t, doc.ReadFromBytes(raw))
	return doc.Root()
}

func encodeElement(t *testing.T, el *etree.Element) string {
	doc := etree.NewDocument()
	doc.SetRoot(el)
	raw, err := doc.WriteToBytes()
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(raw)
}

// unsignedRoot is a LogoutRequest naming the cookie's session, with no signature.
func unsignedRoot(id string) *etree.Element {
	el := etree.NewElement("samlp:LogoutRequest")
	el.CreateAttr("xmlns:samlp", "urn:oasis:names:tc:SAML:2.0:protocol")
	el.CreateAttr("xmlns:saml", "urn:oasis:names:tc:SAML:2.0:assertion")
	el.CreateAttr("ID", id)
	el.CreateAttr("Version", "2.0")
	el.CreateAttr("IssueInstant", time.Now().UTC().Format(time.RFC3339Nano))
	el.CreateAttr("Destination", testSloUrl)
	el.CreateElement("saml:Issuer").SetText(testIdpEntity)
	el.CreateElement("saml:NameID").SetText("rt-user")
	el.CreateElement("samlp:SessionIndex").SetText(testSession)
	return el
}

// A validly signed request for another session must not lend its signature
// to a different element than the one it signs.
func TestServeSloRejectsWrappedSignature(t *testing.T) {
	cases := map[string]func(signed *etree.Element) *etree.Element{
		"signed request nested in an unsigned root": func(signed *etree.Element) *etree.Element {
			root := unsignedRoot("ID_outer")
			root.CreateElement("samlp:Extensions").AddChild(signed)
			return root
		},
		"signature moved onto a root with the signed ID": func(signed *etree.Element) *etree.Element {
			id := signed.SelectAttrValue("ID", "")
			sig := signed.FindElement("./Signature")
			signed.RemoveChild(sig)
			root := unsignedRoot(id)
			root.InsertChildAt(1, sig)
			root.CreateElement("samlp:Extensions").AddChild(signed)
			return root
		},
	}

	for name, wrap := range cases {
		t.Run(name, func(t *testing.T) {
			f := newSloFixture(t)
			signed := signedElement(t, f.logoutRequest(t, func(r *logoutReq) { r.sessionIndex = "user-session-9::client-1" }))
			encoded := encodeElement(t, wrap(signed))

			w := serveSaml(http.MethodPost, "/saml/slo", url.Values{"SAMLRequest": {encoded}}, f.cookie(t, testSession))

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.False(t, deletedCosToken(w), "cos_token must be kept")
			require.False(t, isRevoked(testSession))
			require.False(t, isRevoked("user-session-9::client-1"))
		})
	}
}

func TestServeSamlKeepsAcsPaths(t *testing.T) {
	f := newSloFixture(t)
	encoded := f.logoutRequest(t, nil)

	// A LogoutRequest is not a SAML response: the ACS rejects it as before.
	w := serveSaml(http.MethodPost, "/saml/acs", url.Values{"SAMLRequest": {encoded}}, f.cookie(t, testSession))
	require.Equal(t, http.StatusFound, w.Code)
	require.False(t, deletedCosToken(w))
	require.False(t, isRevoked(testSession))
}

func authRequest(cookie *http.Cookie) (*httptest.ResponseRecorder, bool) {
	reached := false
	router := gin.New()
	router.Use(AuthRequest)
	router.GET("/api/v1/me", func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	r := httptest.NewRequest(http.MethodGet, testSpRoot+"/api/v1/me", nil)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w, reached
}

func TestAuthRequestHonoursSlo(t *testing.T) {
	f := newSloFixture(t)
	cookie := f.cookie(t, testSession)

	w, reached := authRequest(cookie)
	require.True(t, reached)
	require.Equal(t, http.StatusOK, w.Code)

	// The browser that ran the SLO dropped the cookie; a copy of it must stop too.
	encoded := f.logoutRequest(t, nil)
	f.logoutResponse(t, serveSaml(http.MethodPost, "/saml/slo", url.Values{"SAMLRequest": {encoded}}, nil), requestID(t, encoded))

	w, reached = authRequest(cookie)
	require.False(t, reached)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.True(t, deletedCosToken(w))
}

// If Keycloak keeps the session after an SLO (e.g. it never got the
// LogoutResponse), it signs the user in again with the same SessionIndex. The
// new cookie is Keycloak's to judge, or the login loops.
func TestAuthRequestAfterIncompleteSlo(t *testing.T) {
	f := newSloFixture(t)
	clock := &fakeClock{now: time.Now().Add(-2 * time.Minute)}
	prevTimeNow := saml.TimeNow
	saml.TimeNow = clock.Now
	t.Cleanup(func() { saml.TimeNow = prevTimeNow })
	sessionLiveness = newSessionChecker(&fakeLister{ids: map[string]struct{}{"user-session-1": {}}})
	sessionLiveness.now = clock.Now
	old := f.cookie(t, testSession)

	clock.advance(time.Second)
	encoded := f.logoutRequest(t, func(r *logoutReq) { r.issueInstant = clock.now })
	f.logoutResponse(t, serveSaml(http.MethodPost, "/saml/slo", url.Values{"SAMLRequest": {encoded}}, nil), requestID(t, encoded))

	w, reached := authRequest(old)
	require.False(t, reached, "a cookie issued before the SLO is revoked")
	require.Equal(t, http.StatusUnauthorized, w.Code)

	clock.advance(time.Second)
	w, reached = authRequest(f.cookie(t, testSession))
	require.True(t, reached, "a cookie issued after the SLO is left to Keycloak")
	require.Equal(t, http.StatusOK, w.Code)
}

func TestAuthRequestHonoursKeycloakSignOut(t *testing.T) {
	f := newSloFixture(t)
	lister := &fakeLister{ids: map[string]struct{}{"user-session-1": {}}}
	// In the past, so the JWTs issued on this clock stay valid in real time.
	clock := &fakeClock{now: time.Now().Add(-2 * time.Minute)}
	sessionLiveness = newSessionChecker(lister)
	sessionLiveness.now = clock.Now
	prevTimeNow := saml.TimeNow
	saml.TimeNow = clock.Now // cookies are issued on the same clock
	t.Cleanup(func() { saml.TimeNow = prevTimeNow })
	cookie := f.cookie(t, testSession)

	_, reached := authRequest(cookie)
	require.True(t, reached)

	// Admin console "Sign out": Keycloak drops the session.
	lister.set(map[string]struct{}{})
	clock.advance(sessionCheckTTL)

	w, reached := authRequest(cookie)
	require.False(t, reached)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.True(t, deletedCosToken(w))

	// Signing in again gives a new Keycloak session and a new cookie: no loop.
	lister.set(map[string]struct{}{"user-session-2": {}})
	_, reached = authRequest(f.cookie(t, "user-session-2::client-1"))
	require.True(t, reached)
}

type fakeLister struct {
	mu    sync.Mutex
	ids   map[string]struct{}
	err   error
	calls int
}

func (l *fakeLister) ListClientSessionIds(_ context.Context, idOfClient string) (map[string]struct{}, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.calls++
	if l.err != nil {
		return nil, l.err
	}
	if idOfClient != "client-1" {
		return map[string]struct{}{}, nil
	}

	ids := map[string]struct{}{}
	for id := range l.ids {
		ids[id] = struct{}{}
	}
	return ids, nil
}

func (l *fakeLister) set(ids map[string]struct{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ids = ids
}

func (l *fakeLister) fail(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.err = err
}

func (l *fakeLister) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.calls
}

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time          { return c.now }
func (c *fakeClock) advance(d time.Duration) { c.now = c.now.Add(d) }

func TestSessionChecker(t *testing.T) {
	start := time.Unix(1_800_000_000, 0)
	issued := start.Add(-time.Minute)

	newChecker := func() (*sessionChecker, *fakeLister, *fakeClock) {
		lister := &fakeLister{ids: map[string]struct{}{"us-1": {}}}
		clock := &fakeClock{now: start}
		checker := newSessionChecker(lister)
		checker.now = clock.Now
		return checker, lister, clock
	}

	t.Run("cache hit within the TTL", func(t *testing.T) {
		checker, lister, clock := newChecker()
		require.True(t, checker.isAlive("us-1::client-1", issued))
		clock.advance(sessionCheckTTL - time.Second)
		require.True(t, checker.isAlive("us-1::client-1", issued))
		require.Equal(t, 1, lister.count())
	})

	t.Run("revoked within the TTL is caught after it", func(t *testing.T) {
		checker, lister, clock := newChecker()
		require.True(t, checker.isAlive("us-1::client-1", issued))
		lister.set(map[string]struct{}{})
		require.True(t, checker.isAlive("us-1::client-1", issued), "cached until the TTL")
		clock.advance(sessionCheckTTL)
		require.False(t, checker.isAlive("us-1::client-1", issued))
		require.Equal(t, 2, lister.count())
	})

	t.Run("unknown old session is rejected from the cache", func(t *testing.T) {
		checker, lister, _ := newChecker()
		require.True(t, checker.isAlive("us-1::client-1", issued))
		require.False(t, checker.isAlive("us-9::client-1", issued))
		require.Equal(t, 1, lister.count())
	})

	t.Run("session newer than the snapshot forces a refresh", func(t *testing.T) {
		checker, lister, clock := newChecker()
		require.True(t, checker.isAlive("us-1::client-1", issued))
		clock.advance(time.Second)
		lister.set(map[string]struct{}{"us-1": {}, "us-2": {}})
		require.True(t, checker.isAlive("us-2::client-1", clock.now))
		require.Equal(t, 2, lister.count())
	})

	t.Run("keycloak down fails open and backs off", func(t *testing.T) {
		checker, lister, clock := newChecker()
		lister.fail(errors.New("connection refused"))
		require.True(t, checker.isAlive("us-9::client-1", issued))
		require.True(t, checker.isAlive("us-9::client-1", issued))
		require.Equal(t, 1, lister.count(), "no retry within the back-off")
		clock.advance(sessionCheckRetry)
		lister.fail(nil)
		require.False(t, checker.isAlive("us-9::client-1", issued))
		require.Equal(t, 2, lister.count())
	})

	t.Run("not a keycloak session index", func(t *testing.T) {
		checker, lister, _ := newChecker()
		for _, index := range []string{"", "us-1", "us-1::", "::client-1", "a::b::c"} {
			require.True(t, checker.isAlive(index, issued), index)
		}
		require.Equal(t, 0, lister.count())
	})
}
