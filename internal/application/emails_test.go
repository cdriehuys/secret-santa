package application_test

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/cdriehuys/secret-santa/internal/application"
	"github.com/cdriehuys/secret-santa/internal/application/testutils"
)

const (
	expectedVerificationPathSegment = "verify-email"
)

type capturingMailer struct {
	sendTo      string
	sendFrom    string
	sendSubject string
	sendBody    string
	sendError   error
}

func (m *capturingMailer) Send(ctx context.Context, to string, from string, subject string, body string) error {
	m.sendTo = to
	m.sendFrom = from
	m.sendSubject = subject
	m.sendBody = body

	return m.sendError
}

func TestEmailVerifier_NewEmail(t *testing.T) {
	testCases := []struct {
		name       string
		mailer     capturingMailer
		templates  testutils.CapturingTemplateData
		baseDomain string
		sender     string
		email      string
		token      string
		wantErr    bool
	}{
		{
			name:       "successful send",
			baseDomain: "http://localhost",
			sender:     "admin@localhost",
			email:      "new-user@example.com",
			token:      "secret-token",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			baseDomain, err := url.Parse(tt.baseDomain)
			if err != nil {
				t.Fatalf("Base domain %q is invalid: %v", tt.baseDomain, err)
			}

			verifier := application.NewEmailVerifier(slog.New(slog.DiscardHandler), &tt.mailer, nil, baseDomain, tt.sender)

			err = verifier.NewEmail(t.Context(), tt.email, tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error presence %v, got error %#v", tt.wantErr, err)
			}

			if got := tt.mailer.sendTo; got != tt.email {
				t.Errorf("Expected email to be sent to %q, got %q", tt.email, got)
			}

			if got := tt.mailer.sendFrom; got != tt.sender {
				t.Errorf("Expected email to be sent from %q, got %q", tt.sender, got)
			}

			if got := tt.mailer.sendSubject; got == "" {
				t.Error("Expected non-empty subject.")
			}

			wantLink := baseDomain.JoinPath(expectedVerificationPathSegment, tt.token).String()
			if !strings.Contains(tt.mailer.sendBody, wantLink) {
				t.Errorf("Expected verification link %q to be present in body:\n%s", wantLink, tt.mailer.sendBody)
			}
		})
	}
}
