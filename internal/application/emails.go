package application

import (
	"context"
	"log/slog"
	"net/url"
)

type Emailer interface {
	Send(ctx context.Context, to string, from string, subject string, body string) error
}

type EmailVerifier struct {
	logger *slog.Logger

	emailer   Emailer
	templates TemplateEngine

	baseDomain *url.URL
	sender     string
}

func NewEmailVerifier(logger *slog.Logger, emailer Emailer, templates TemplateEngine, baseDomain *url.URL, sender string) *EmailVerifier {
	return &EmailVerifier{
		logger:     logger,
		emailer:    emailer,
		templates:  templates,
		baseDomain: baseDomain,
		sender:     sender,
	}
}

func (e *EmailVerifier) DuplicateRegistration(ctx context.Context, email string) error {
	panic("unimplemented")
}

func (v *EmailVerifier) NewEmail(ctx context.Context, email string, token string) error {
	verificationLink := v.baseDomain.JoinPath("verify-email", token).String()

	return v.emailer.Send(ctx, email, v.sender, "Verify Your Email", verificationLink)
}
