package saml

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/xmlenc"
	xrv "github.com/mattermost/xml-roundtrip-validator"
	"github.com/stretchr/testify/require"
)

const (
	testDigestSHA1   = "http://www.w3.org/2000/09/xmldsig#sha1"
	testDigestSHA256 = "http://www.w3.org/2001/04/xmlenc#sha256"
	testMGF1SHA1     = "http://www.w3.org/2009/xmlenc11#mgf1sha1"
	testMGF1SHA256   = "http://www.w3.org/2009/xmlenc11#mgf1sha256"
	testRSAOAEPMGF1P = "http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p"
)

var testAssertion = []byte(`<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="_a1"/>`)

// sealed is one EncryptedAssertion's worth of inputs, laid out the way Keycloak 26.7.5
// sends it: EncryptedData > KeyInfo > EncryptedKey, the digest and MGF as children of
// the key's EncryptionMethod. An empty digest or mgf leaves that element out; a
// non-empty oaep.Label is sent as OAEPparams; spaced breaks every base64 value with
// the XML whitespace xs:base64Binary allows.
type sealed struct {
	cipher    string
	keySize   int
	transport string
	digest    string
	mgf       string
	oaep      rsa.OAEPOptions
	spaced    bool
}

func (s sealed) base64(b []byte) string {
	text := base64.StdEncoding.EncodeToString(b)
	if !s.spaced {
		return text
	}
	var out strings.Builder
	for i, r := range text {
		if i > 0 && i%8 == 0 {
			out.WriteString([]string{" ", "\t", "\n  "}[i/8%3])
		}
		out.WriteRune(r)
	}
	return out.String()
}

func (s sealed) encryptedData(t *testing.T, pub *rsa.PublicKey, plaintext []byte) *etree.Element {
	t.Helper()

	cek := make([]byte, s.keySize)
	_, err := rand.Read(cek)
	require.NoError(t, err)
	block, err := aes.NewCipher(cek)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	require.NoError(t, err)
	body := gcm.Seal(nonce, nonce, plaintext, nil)

	wrapped, err := rsa.EncryptOAEPWithOptions(rand.Reader, pub, cek, &s.oaep)
	require.NoError(t, err)

	doc := etree.NewDocument()
	ed := doc.CreateElement("xenc:EncryptedData")
	ed.CreateAttr("xmlns:xenc", "http://www.w3.org/2001/04/xmlenc#")
	ed.CreateAttr("Type", "http://www.w3.org/2001/04/xmlenc#Element")
	ed.CreateElement("xenc:EncryptionMethod").CreateAttr("Algorithm", s.cipher)
	ki := ed.CreateElement("ds:KeyInfo")
	ki.CreateAttr("xmlns:ds", "http://www.w3.org/2000/09/xmldsig#")
	ek := ki.CreateElement("xenc:EncryptedKey")
	em := ek.CreateElement("xenc:EncryptionMethod")
	em.CreateAttr("Algorithm", s.transport)
	if s.digest != "" {
		em.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", s.digest)
	}
	if s.mgf != "" {
		mgf := em.CreateElement("xenc11:MGF")
		mgf.CreateAttr("xmlns:xenc11", "http://www.w3.org/2009/xmlenc11#")
		mgf.CreateAttr("Algorithm", s.mgf)
	}
	if len(s.oaep.Label) > 0 {
		em.CreateElement("xenc:OAEPparams").SetText(s.base64(s.oaep.Label))
	}
	ek.CreateElement("xenc:CipherData").CreateElement("xenc:CipherValue").SetText(s.base64(wrapped))
	ed.CreateElement("xenc:CipherData").CreateElement("xenc:CipherValue").SetText(s.base64(body))
	return ed
}

func testKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

// The same call crewjam's decryptElement makes when the EncryptedKey sits in KeyInfo.
func TestDecryptKeycloakEncryptedAssertion(t *testing.T) {
	key := testKey(t)
	cases := map[string]sealed{
		// Keycloak 26 with the cipher left unset: a fresh 3.2.0 install.
		"aes256-gcm, xmlenc11#rsa-oaep sha256/mgf1sha256": {
			cipher: xmlenc11AES256GCM, keySize: 32, transport: xmlenc11RSAOAEP,
			digest: testDigestSHA256, mgf: testMGF1SHA256,
			oaep: rsa.OAEPOptions{Hash: crypto.SHA256, MGFHash: crypto.SHA256},
		},
		// An upgraded cluster after terraform drops the AES-128-CBC pin: the key
		// transport stays as migrated, the cipher falls back to the default.
		"aes256-gcm, rsa-oaep-mgf1p sha1": {
			cipher: xmlenc11AES256GCM, keySize: 32, transport: testRSAOAEPMGF1P,
			digest: testDigestSHA1,
			oaep:   rsa.OAEPOptions{Hash: crypto.SHA1, MGFHash: crypto.SHA1},
		},
		"aes192-gcm, xmlenc11#rsa-oaep": {
			cipher: xmlenc11AES192GCM, keySize: 24, transport: xmlenc11RSAOAEP,
			digest: testDigestSHA256, mgf: testMGF1SHA256,
			oaep: rsa.OAEPOptions{Hash: crypto.SHA256, MGFHash: crypto.SHA256},
		},
		"xmlenc11#rsa-oaep, digest given, MGF defaults to mgf1sha1": {
			cipher: xmlenc11AES256GCM, keySize: 32, transport: xmlenc11RSAOAEP,
			digest: testDigestSHA256,
			oaep:   rsa.OAEPOptions{Hash: crypto.SHA256, MGFHash: crypto.SHA1},
		},
		"xmlenc11#rsa-oaep, digest and MGF default to SHA-1": {
			cipher: xmlenc11AES256GCM, keySize: 32, transport: xmlenc11RSAOAEP,
			oaep: rsa.OAEPOptions{Hash: crypto.SHA1, MGFHash: crypto.SHA1},
		},
		"xmlenc11#rsa-oaep, explicit mgf1sha1": {
			cipher: xmlenc11AES256GCM, keySize: 32, transport: xmlenc11RSAOAEP,
			digest: testDigestSHA256, mgf: testMGF1SHA1,
			oaep: rsa.OAEPOptions{Hash: crypto.SHA256, MGFHash: crypto.SHA1},
		},
		"OAEPparams label, base64 broken by spaces, tabs and newlines": {
			cipher: xmlenc11AES256GCM, keySize: 32, transport: xmlenc11RSAOAEP,
			digest: testDigestSHA256, mgf: testMGF1SHA256,
			oaep:   rsa.OAEPOptions{Hash: crypto.SHA256, MGFHash: crypto.SHA256, Label: []byte("cube-cos-api")},
			spaced: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ed := c.encryptedData(t, &key.PublicKey, testAssertion)
			plaintext, err := xmlenc.Decrypt(key, ed)
			require.NoError(t, err)
			require.Equal(t, testAssertion, plaintext)
		})
	}
}

func TestDecryptKeycloakEncryptedAssertionRejects(t *testing.T) {
	key := testKey(t)
	keycloak := sealed{
		cipher: xmlenc11AES256GCM, keySize: 32, transport: xmlenc11RSAOAEP,
		digest: testDigestSHA256, mgf: testMGF1SHA256,
		oaep: rsa.OAEPOptions{Hash: crypto.SHA256, MGFHash: crypto.SHA256},
	}

	t.Run("tampered ciphertext", func(t *testing.T) {
		ed := keycloak.encryptedData(t, &key.PublicKey, testAssertion)
		value := ed.FindElement("./CipherData/CipherValue")
		body, err := base64.StdEncoding.DecodeString(value.Text())
		require.NoError(t, err)
		body[len(body)-1] ^= 0xff
		value.SetText(base64.StdEncoding.EncodeToString(body))

		_, err = xmlenc.Decrypt(key, ed)
		require.Error(t, err)
	})

	t.Run("another key", func(t *testing.T) {
		ed := keycloak.encryptedData(t, &key.PublicKey, testAssertion)
		_, err := xmlenc.Decrypt(testKey(t), ed)
		require.Error(t, err)
	})

	t.Run("MGF and digest mismatch", func(t *testing.T) {
		wrong := keycloak
		wrong.oaep.MGFHash = crypto.SHA1 // the element still says mgf1sha256
		ed := wrong.encryptedData(t, &key.PublicKey, testAssertion)
		_, err := xmlenc.Decrypt(key, ed)
		require.Error(t, err)
	})

	t.Run("OAEPparams label differs", func(t *testing.T) {
		labelled := keycloak
		labelled.oaep.Label = []byte("cube-cos-api")
		ed := labelled.encryptedData(t, &key.PublicKey, testAssertion)
		ed.FindElement("./KeyInfo/EncryptedKey/EncryptionMethod/OAEPparams").SetText(base64.StdEncoding.EncodeToString([]byte("other")))
		_, err := xmlenc.Decrypt(key, ed)
		require.Error(t, err)
	})

	t.Run("ciphertext shorter than IV and tag", func(t *testing.T) {
		ed := keycloak.encryptedData(t, &key.PublicKey, testAssertion)
		ed.FindElement("./CipherData/CipherValue").SetText(base64.StdEncoding.EncodeToString(make([]byte, gcmNonceSize+gcmTagSize-1)))
		_, err := xmlenc.Decrypt(key, ed)
		require.ErrorContains(t, err, "shorter than IV and tag")
	})

	t.Run("unknown MGF", func(t *testing.T) {
		unknown := keycloak
		unknown.mgf = "http://example.com/mgf1md5"
		ed := unknown.encryptedData(t, &key.PublicKey, testAssertion)
		_, err := xmlenc.Decrypt(key, ed)
		require.ErrorIs(t, err, xmlenc.ErrAlgorithmNotImplemented(unknown.mgf))
	})

	t.Run("key size does not match the cipher", func(t *testing.T) {
		short := keycloak
		short.keySize = 16 // AES-128 key under an aes256-gcm label
		ed := short.encryptedData(t, &key.PublicKey, testAssertion)
		_, err := xmlenc.Decrypt(key, ed)
		require.ErrorContains(t, err, "expected a 32-byte key")
	})
}

func TestSamlFailureReason(t *testing.T) {
	cause := errors.New("algorithm is not implemented")
	wrapped := &saml.InvalidResponseError{PrivateErr: cause, Response: "<samlp:Response>SECRET-RESPONSE</samlp:Response>"}

	require.Equal(t, `"algorithm is not implemented"`, samlFailureReason(wrapped))
	require.Equal(t, `"algorithm is not implemented"`, samlFailureReason(fmt.Errorf("parse: %w", wrapped)))
	require.NotContains(t, samlFailureReason(wrapped), "SECRET-RESPONSE")
	require.Equal(t, `"no tracked request"`, samlFailureReason(errors.New("no tracked request")))

	t.Run("XML round-trip detail is withheld", func(t *testing.T) {
		// What crewjam's decryptElement returns when the decrypted assertion fails xrv.
		roundtrip := xrv.XMLRoundtripError{
			Expected: xml.StartElement{Name: xml.Name{Local: "Assertion"}, Attr: []xml.Attr{{Name: xml.Name{Local: "ID"}, Value: "SECRET-ASSERTION"}}},
			Observed: xml.StartElement{Name: xml.Name{Local: "Assertion"}},
		}
		require.Contains(t, roundtrip.Error(), "SECRET-ASSERTION")
		err := &saml.InvalidResponseError{PrivateErr: fmt.Errorf("failed to decrypt EncryptedAssertion: %v",
			fmt.Errorf("plaintext response contains invalid XML: %s", roundtrip))}

		reason := samlFailureReason(err)
		require.NotContains(t, reason, "SECRET-ASSERTION")
		require.Contains(t, reason, "details withheld")
	})

	t.Run("response-controlled text is quoted and bounded", func(t *testing.T) {
		injected := errors.New("`Destination` does not match: https://evil\n2026-10-06 INFO forged line" + strings.Repeat("x", 2*maxFailureReasonLen))
		reason := samlFailureReason(&saml.InvalidResponseError{PrivateErr: injected})
		require.NotContains(t, reason, "\n2026")
		require.Contains(t, reason, `\n2026`)
		require.Less(t, len(reason), maxFailureReasonLen+16)
	})
}
