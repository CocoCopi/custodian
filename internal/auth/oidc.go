package auth

import (
	"context"
	"fmt"
	"time"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider wraps the OpenID Connect discovery + verification flow.
type OIDCProvider struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    oauth2.Config
}

// NewOIDCProvider performs OIDC discovery against the issuer and builds an
// OAuth2 config for the authorization-code flow.
func NewOIDCProvider(ctx context.Context, issuer, clientID, clientSecret, redirectURL string) (*OIDCProvider, error) {
	if issuer == "" {
		return nil, nil // OIDC disabled
	}
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery for %s: %w", issuer, err)
	}
	return &OIDCProvider{
		provider: provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
		oauth: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  redirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}, nil
}

// AuthCodeURL returns the redirect URL to start the login flow.
func (p *OIDCProvider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// Exchange verifies the authorization code and returns verified ID token claims.
func (p *OIDCProvider) Exchange(ctx context.Context, code string) (subject, name, email string, err error) {
	oauthToken, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return "", "", "", fmt.Errorf("oidc exchange: %w", err)
	}
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return "", "", "", fmt.Errorf("no id_token in OIDC response")
	}
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", "", "", fmt.Errorf("oidc verify: %w", err)
	}
	var claims struct {
		Subject string `json:"sub"`
		Name    string `json:"name"`
		Email   string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", "", "", err
	}
	return claims.Subject, claims.Name, claims.Email, nil
}

// Enabled reports whether OIDC is configured.
func (p *OIDCProvider) Enabled() bool { return p != nil && p.provider != nil }

// Now is a small indirection so tests can pin time.
var Now = time.Now
