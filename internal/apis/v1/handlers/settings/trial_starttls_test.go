package settings

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	emailv2 "github.com/bigstack-oss/bigstack-dependency-go/pkg/email/v2"
	emailv1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/stretchr/testify/require"
)

// selfSignedCert is a certificate no trust store knows, like a relay with a
// self-signed certificate or one from an internal CA.
func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "relay.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// startTLSSession is what the fake relay saw on its single connection.
type startTLSSession struct {
	tls       bool   // upgraded with STARTTLS
	authInTLS bool   // AUTH arrived after the upgrade
	authPlain bool   // AUTH arrived before any upgrade
	payload   string // message body, empty when nothing arrived
}

// startTLSRelay offers STARTTLS with a self-signed certificate and AUTH PLAIN,
// like an internal relay behind an internal CA.
type startTLSRelay struct {
	host string
	port int
	done chan startTLSSession
}

func newStartTLSRelay(t *testing.T) *startTLSRelay {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	r := &startTLSRelay{host: host, port: port, done: make(chan startTLSSession, 1)}
	cfg := &tls.Config{Certificates: []tls.Certificate{selfSignedCert(t)}}
	go r.serve(ln, cfg)
	return r
}

func (r *startTLSRelay) serve(ln net.Listener, cfg *tls.Config) {
	var s startTLSSession
	defer func() { r.done <- s }()

	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	rd := bufio.NewReader(conn)
	w := func(line string) { _, _ = conn.Write([]byte(line + "\r\n")) }
	w("220 relay.test ESMTP ready")
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			return
		}

		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			w("250-relay.test")
			if !s.tls {
				w("250-STARTTLS")
			}
			w("250 AUTH PLAIN")
		case cmd == "STARTTLS":
			w("220 2.0.0 Ready to start TLS")
			tlsConn := tls.Server(conn, cfg)
			if err := tlsConn.Handshake(); err != nil {
				// the client refused the certificate
				return
			}
			conn = tlsConn
			rd = bufio.NewReader(conn)
			s.tls = true
		case strings.HasPrefix(cmd, "AUTH"):
			if s.tls {
				s.authInTLS = true
			} else {
				s.authPlain = true
			}
			w("235 2.7.0 Authentication successful")
		case strings.HasPrefix(cmd, "DATA"):
			w("354 End data with <CR><LF>.<CR><LF>")
			var b strings.Builder
			for {
				dl, err := rd.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimRight(dl, "\r\n") == "." {
					break
				}
				b.WriteString(dl)
			}
			s.payload = b.String()
			w("250 2.0.0 OK: queued")
		case strings.HasPrefix(cmd, "QUIT"):
			w("221 2.0.0 Bye")
			return
		default:
			w("250 2.0.0 OK")
		}
	}
}

// TestSendTrialEmailCertificateRule pins that the test email follows the alert
// path's certificate rule (cubecos config_kapacitor.cpp: no-verify unless the
// policy is mandatory) against a relay with a self-signed certificate.
func TestSendTrialEmailCertificateRule(t *testing.T) {
	tests := []struct {
		name     string
		tls      emailv2.TLS
		wantSent bool
	}{
		{name: "opportunistic accepts a certificate it cannot verify, as alert mail does", tls: emailv2.TLSOpportunistic, wantSent: true},
		{name: "mandatory still verifies the certificate", tls: emailv2.TLSMandatory, wantSent: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relay := newStartTLSRelay(t)
			auth := true
			password := "secret"
			sender := &emailv1.Sender{
				Host:     relay.host,
				Port:     relay.port,
				Auth:     &auth,
				Username: "user",
				Password: &password,
				TLS:      tt.tls,
				From:     "noreply@example.com",
			}

			err := sendTrialEmail(sender, "admin@example.com")
			s := <-relay.done
			require.False(t, s.authPlain, "credentials must never be sent before STARTTLS")

			if !tt.wantSent {
				require.Error(t, err)
				require.False(t, s.authInTLS, "credentials must not reach a relay whose certificate failed")
				require.Empty(t, s.payload)
				return
			}

			require.NoError(t, err)
			require.True(t, s.tls, "expected a STARTTLS upgrade")
			require.True(t, s.authInTLS, "expected AUTH inside TLS")
			require.Contains(t, s.payload, "Subject: Test Email")
		})
	}
}

// fakeHelper stands in for a future pkg/email/v2 that no longer hands back a
// *emailv2.Helper holding a go-mail client.
type fakeHelper struct{}

func (fakeHelper) SetClient() error           { return nil }
func (fakeHelper) Send(emailv2.Message) error { return nil }

// TestSkipVerifyUnlessMandatoryFailsLoudly guards the type assertions: when
// pkg/email/v2 changes shape, the test email must fail rather than silently
// keep verifying (or skip) the certificate.
func TestSkipVerifyUnlessMandatoryFailsLoudly(t *testing.T) {
	sender := &emailv1.Sender{Host: "relay.test", TLS: emailv2.TLSOpportunistic}
	require.Error(t, skipVerifyUnlessMandatory(fakeHelper{}, sender))

	// policies other than opportunistic leave the helper alone
	sender.TLS = emailv2.TLSMandatory
	require.NoError(t, skipVerifyUnlessMandatory(fakeHelper{}, sender))
	sender.TLS = emailv2.TLSNone
	require.NoError(t, skipVerifyUnlessMandatory(fakeHelper{}, sender))
}
