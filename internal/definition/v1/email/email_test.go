package email

import (
	"testing"

	emailv2 "github.com/bigstack-oss/bigstack-dependency-go/pkg/email/v2"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func TestSenderNormalize(t *testing.T) {
	tests := []struct {
		name         string
		sender       Sender
		current      *Sender
		wantAuth     bool
		wantTLS      emailv2.TLS
		wantUsername string
		wantPassword string
		wantErr      bool
	}{
		{
			name:         "unset auth with a username authenticates",
			sender:       Sender{Username: "user", Password: strPtr("secret")},
			wantAuth:     true,
			wantTLS:      emailv2.TLSMandatory,
			wantUsername: "user",
			wantPassword: "secret",
		},
		{
			name:     "unset auth without a username is anonymous",
			sender:   Sender{},
			wantAuth: false,
			wantTLS:  emailv2.TLSMandatory,
		},
		{
			name:     "auth off drops the credentials",
			sender:   Sender{Auth: boolPtr(false), Username: "user", Password: strPtr("secret"), TLS: emailv2.TLSNone},
			wantAuth: false,
			wantTLS:  emailv2.TLSNone,
		},
		{
			name:         "unset fields keep the current sender's",
			sender:       Sender{Username: "user", Password: strPtr("secret")},
			current:      &Sender{Auth: boolPtr(false), TLS: emailv2.TLSOpportunistic},
			wantAuth:     false,
			wantTLS:      emailv2.TLSOpportunistic,
			wantUsername: "",
			wantPassword: "",
		},
		{
			name:         "explicit fields override the current sender's",
			sender:       Sender{Auth: boolPtr(true), Username: "user", Password: strPtr("secret"), TLS: emailv2.TLSMandatory},
			current:      &Sender{Auth: boolPtr(false), TLS: emailv2.TLSNone},
			wantAuth:     true,
			wantTLS:      emailv2.TLSMandatory,
			wantUsername: "user",
			wantPassword: "secret",
		},
		{
			name:     "opportunistic is accepted",
			sender:   Sender{Auth: boolPtr(false), TLS: emailv2.TLSOpportunistic},
			wantAuth: false,
			wantTLS:  emailv2.TLSOpportunistic,
		},
		{
			name:    "an unknown tls is rejected",
			sender:  Sender{TLS: "starttls"},
			wantErr: true,
		},
		{
			name:    "tls is case sensitive",
			sender:  Sender{TLS: "None"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.sender
			err := s.Normalize(tt.current)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantAuth, s.IsAuth())
			require.Equal(t, tt.wantTLS, s.TLS)
			require.Equal(t, tt.wantUsername, s.Username)
			require.Equal(t, tt.wantPassword, s.GetPassword())
		})
	}
}

func TestSenderToHexSchemaNeverNullPassword(t *testing.T) {
	s := Sender{Host: "relay", Port: 25, Auth: boolPtr(false), TLS: emailv2.TLSNone, From: "a@b.co"}
	require.Equal(t, HexSender{Host: "relay", Port: 25, Auth: false, Password: "", TLS: "none", From: "a@b.co"}, s.ToHexSchema())
}

func TestCosSenderToApiSchema(t *testing.T) {
	tests := []struct {
		name     string
		cos      CosSender
		wantAuth bool
		wantTLS  emailv2.TLS
	}{
		{
			name:     "a sender stored before tls existed stays opportunistic",
			cos:      CosSender{Host: "relay", Port: "25", Username: "user", Auth: true},
			wantAuth: true,
			wantTLS:  emailv2.TLSOpportunistic,
		},
		{
			name:     "an anonymous sender without tls",
			cos:      CosSender{Host: "relay", Port: "25", Auth: false, TLS: "none"},
			wantAuth: false,
			wantTLS:  emailv2.TLSNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.cos.ToApiSchema()
			require.Equal(t, tt.wantAuth, s.IsAuth())
			require.Equal(t, tt.wantTLS, s.TLS)
			require.Equal(t, 25, s.Port)
		})
	}
}
