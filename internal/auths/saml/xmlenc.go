package saml

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	_ "crypto/sha1" //nolint:gosec // SHA-1 is the XML Encryption default for OAEP's digest and MGF1
	_ "crypto/sha256"
	_ "crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/crewjam/saml/xmlenc"
)

// Keycloak 26 encrypts assertions with XML Encryption 1.1: the key is wrapped with
// xmlenc11#rsa-oaep and the assertion is sealed with AES-256-GCM. crewjam/saml v0.5.1,
// its latest release, registers neither, so ParseResponse rejects every encrypted
// assertion with "algorithm is not implemented" and no session is created.
// Registering them here lets the ACS decrypt what Keycloak sends by default.
const (
	xmlenc11RSAOAEP   = "http://www.w3.org/2009/xmlenc11#rsa-oaep"
	xmlenc11AES192GCM = "http://www.w3.org/2009/xmlenc11#aes192-gcm"
	xmlenc11AES256GCM = "http://www.w3.org/2009/xmlenc11#aes256-gcm"

	gcmNonceSize = 12
	gcmTagSize   = 16
)

// The DigestMethod URIs XML Encryption allows for RSA-OAEP. Keycloak writes
// xmlenc#sha256, which crewjam's digest registry (xmldsig URIs only) lacks.
var oaepDigests = map[string]crypto.Hash{
	"http://www.w3.org/2000/09/xmldsig#sha1":        crypto.SHA1,
	"http://www.w3.org/2001/04/xmlenc#sha256":       crypto.SHA256,
	"http://www.w3.org/2001/04/xmldsig-more#sha384": crypto.SHA384,
	"http://www.w3.org/2001/04/xmlenc#sha512":       crypto.SHA512,
}

// The MGF URIs XML Encryption 1.1 defines for xmlenc11#rsa-oaep.
var mgfHashes = map[string]crypto.Hash{
	"http://www.w3.org/2009/xmlenc11#mgf1sha1":   crypto.SHA1,
	"http://www.w3.org/2009/xmlenc11#mgf1sha224": crypto.SHA224,
	"http://www.w3.org/2009/xmlenc11#mgf1sha256": crypto.SHA256,
	"http://www.w3.org/2009/xmlenc11#mgf1sha384": crypto.SHA384,
	"http://www.w3.org/2009/xmlenc11#mgf1sha512": crypto.SHA512,
}

func init() {
	xmlenc.RegisterDecrypter(rsaOAEP11{})
	xmlenc.RegisterDecrypter(aesGCM{algorithm: xmlenc11AES192GCM, keySize: 24})
	xmlenc.RegisterDecrypter(aesGCM{algorithm: xmlenc11AES256GCM, keySize: 32})
}

// rsaOAEP11 unwraps an EncryptedKey sent as xmlenc11#rsa-oaep. Unlike the 2001
// rsa-oaep-mgf1p, the MGF1 hash is named separately from the digest; both
// default to SHA-1 when the element is absent.
type rsaOAEP11 struct{}

func (rsaOAEP11) Algorithm() string {
	return xmlenc11RSAOAEP
}

func (rsaOAEP11) Decrypt(key interface{}, encryptedKeyEl *etree.Element) ([]byte, error) {
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("xmlenc11#rsa-oaep: expected key to be a *rsa.PrivateKey")
	}

	digest, err := lookupHash(encryptedKeyEl, "./EncryptionMethod/DigestMethod", oaepDigests)
	if err != nil {
		return nil, err
	}
	mgf, err := lookupHash(encryptedKeyEl, "./EncryptionMethod/MGF", mgfHashes)
	if err != nil {
		return nil, err
	}

	var label []byte
	if el := encryptedKeyEl.FindElement("./EncryptionMethod/OAEPparams"); el != nil {
		label, err = decodeBase64Binary(el.Text())
		if err != nil {
			return nil, fmt.Errorf("xmlenc11#rsa-oaep: OAEPparams: %w", err)
		}
	}

	ciphertext, err := cipherValue(encryptedKeyEl)
	if err != nil {
		return nil, err
	}
	return rsaKey.Decrypt(nil, ciphertext, &rsa.OAEPOptions{Hash: digest, MGFHash: mgf, Label: label})
}

// aesGCM opens EncryptedData sealed with AES-GCM as XML Encryption 1.1 lays it out:
// a 96-bit IV, then the ciphertext, then the 128-bit tag.
type aesGCM struct {
	algorithm string
	keySize   int
}

func (g aesGCM) Algorithm() string {
	return g.algorithm
}

func (g aesGCM) Decrypt(key interface{}, encryptedDataEl *etree.Element) ([]byte, error) {
	// The wrapped key may sit inside the EncryptedData, as crewjam's own ciphers allow.
	if encryptedKeyEl := encryptedDataEl.FindElement("./KeyInfo/EncryptedKey"); encryptedKeyEl != nil {
		var err error
		key, err = xmlenc.Decrypt(key, encryptedKeyEl)
		if err != nil {
			return nil, err
		}
	}

	keyBuf, ok := key.([]byte)
	if !ok {
		return nil, fmt.Errorf("%s: expected a []byte key", g.algorithm)
	}
	if len(keyBuf) != g.keySize {
		return nil, fmt.Errorf("%s: expected a %d-byte key, got %d", g.algorithm, g.keySize, len(keyBuf))
	}

	ciphertext, err := cipherValue(encryptedDataEl)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcmNonceSize+gcmTagSize {
		return nil, fmt.Errorf("%s: ciphertext is %d bytes, shorter than IV and tag", g.algorithm, len(ciphertext))
	}

	block, err := aes.NewCipher(keyBuf)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, ciphertext[:gcmNonceSize], ciphertext[gcmNonceSize:], nil)
}

// lookupHash reads the Algorithm of the element at path, SHA-1 when it is absent.
func lookupHash(el *etree.Element, path string, known map[string]crypto.Hash) (crypto.Hash, error) {
	methodEl := el.FindElement(path)
	if methodEl == nil {
		return crypto.SHA1, nil
	}
	uri := methodEl.SelectAttrValue("Algorithm", "")
	h, ok := known[uri]
	if !ok {
		return 0, xmlenc.ErrAlgorithmNotImplemented(uri)
	}
	return h, nil
}

func cipherValue(el *etree.Element) ([]byte, error) {
	valueEl := el.FindElement("./CipherData/CipherValue")
	if valueEl == nil {
		return nil, errors.New("cannot find CipherData/CipherValue")
	}
	return decodeBase64Binary(valueEl.Text())
}

// decodeBase64Binary decodes xs:base64Binary, which may carry XML whitespace
// anywhere; Go's decoder only skips CR and LF.
func decodeBase64Binary(text string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(strings.Join(strings.Fields(text), ""))
}
