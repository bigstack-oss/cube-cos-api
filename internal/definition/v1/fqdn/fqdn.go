package fqdn

import (
	"net"
	"regexp"
	"strings"
)

const (
	Module = "fqdn"
	// Tuning is the hex_config tuning that holds the FQDN on every node.
	Tuning = "cubesys.fqdn"

	// CertificateDir is where hex_sdk fqdn_cert_install puts the certificate on every control.
	CertificateDir = "/etc/cube/certs"
	// the files hex_sdk fqdn_cert_verify and fqdn_cert_install read from a directory
	CertificateFile   = "server.cert"
	PrivateKeyFile    = "server.key"
	CaCertificateFile = "ca.cert"
)

// the pattern hex_config validates cubesys.fqdn with (config_cubesys.cpp): at least two labels,
// the last starting with a letter, so an IP address is not mistaken for a name
var nameRegex = regexp.MustCompile(`^([A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?\.)+[A-Za-z]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

type SetReq struct {
	Fqdn string `json:"fqdn"`
}

type CertificateReq struct {
	// the certificate, optionally followed by its intermediates (PEM)
	Certificate string `json:"certificate"`
	// its private key, unencrypted (PEM)
	PrivateKey string `json:"privateKey"`
	// the root CA that issued it, optionally with the intermediates (PEM)
	CaCertificate string `json:"caCertificate"`
}

type Certificate struct {
	Subject     string   `json:"subject"`
	Issuer      string   `json:"issuer"`
	DnsNames    []string `json:"dnsNames"`
	NotBefore   string   `json:"notBefore"`
	NotAfter    string   `json:"notAfter"`
	Fingerprint string   `json:"fingerprint"`
}

// Node is one node's view: the FQDN its settings carry, the host it sends browsers to, and
// which certificate its browser-facing listeners serve.
type Node struct {
	Hostname    string       `json:"hostname"`
	Role        string       `json:"role,omitempty"`
	Fqdn        string       `json:"fqdn"`
	PublicHost  string       `json:"publicHost"`
	ServedFrom  string       `json:"servedFrom"`
	Certificate *Certificate `json:"certificate,omitempty"`
	Error       string       `json:"error,omitempty"`
}

type Status struct {
	Fqdn        string       `json:"fqdn"`
	PublicHost  string       `json:"publicHost"`
	Certificate *Certificate `json:"certificate"`
	Nodes       []Node       `json:"nodes"`
}

type Result struct {
	Fqdn   string `json:"fqdn"`
	Output string `json:"output"`
}

// Normalize trims and lowercases a name: DNS compares names case-insensitively, and the
// certificate check and every URL then see one spelling.
func Normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func IsValid(name string) bool {
	if len(name) > 253 || net.ParseIP(name) != nil {
		return false
	}

	return nameRegex.MatchString(name)
}

func (c *Certificate) IsEmpty() bool {
	return c == nil || c.Fingerprint == ""
}
