package settings

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"testing"

	emailv2 "github.com/bigstack-oss/bigstack-dependency-go/pkg/email/v2"
	emailv1 "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/email"
	"github.com/stretchr/testify/require"
)

// fakeRelay is a plaintext SMTP relay that advertises AUTH PLAIN but not
// STARTTLS, like an internal relay that accepts anonymous submission.
type fakeRelay struct {
	host     string
	port     int
	sawAuth  chan bool
	received chan string
}

func newFakeRelay(t *testing.T) *fakeRelay {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	r := &fakeRelay{host: host, port: port, sawAuth: make(chan bool, 1), received: make(chan string, 1)}
	go r.serve(ln)
	return r
}

func (r *fakeRelay) serve(ln net.Listener) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	auth := false
	defer func() { r.sawAuth <- auth }()

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
			w("250 AUTH PLAIN")
		case strings.HasPrefix(cmd, "AUTH"):
			auth = true
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
			r.received <- b.String()
			w("250 2.0.0 OK: queued")
		case strings.HasPrefix(cmd, "QUIT"):
			w("221 2.0.0 Bye")
			return
		default:
			w("250 2.0.0 OK")
		}
	}
}

func TestSendTrialEmail(t *testing.T) {
	tests := []struct {
		name     string
		auth     bool
		tls      emailv2.TLS
		wantAuth bool
		wantErr  bool
	}{
		{
			name:     "anonymous relay without tls never attempts auth",
			auth:     false,
			tls:      emailv2.TLSNone,
			wantAuth: false,
		},
		{
			name:     "opportunistic stays plaintext when starttls is not offered",
			auth:     false,
			tls:      emailv2.TLSOpportunistic,
			wantAuth: false,
		},
		{
			name:     "auth over a plaintext link authenticates",
			auth:     true,
			tls:      emailv2.TLSNone,
			wantAuth: true,
		},
		{
			name:    "mandatory fails when starttls is not offered",
			auth:    false,
			tls:     emailv2.TLSMandatory,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relay := newFakeRelay(t)
			password := "secret"
			sender := &emailv1.Sender{
				Host:     relay.host,
				Port:     relay.port,
				Auth:     &tt.auth,
				Username: "user",
				Password: &password,
				TLS:      tt.tls,
				From:     "noreply@example.com",
			}

			err := sendTrialEmail(sender, "admin@example.com")
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Contains(t, <-relay.received, "Subject: Test Email")
			require.Equal(t, tt.wantAuth, <-relay.sawAuth)
		})
	}
}
