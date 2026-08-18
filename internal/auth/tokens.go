// Package auth implements authentication for the control plane: OIDC
// (Keycloak / Authelia / any OpenID Connect provider) for interactive
// sessions and long-lived API tokens for CLI and CI usage.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService creates, verifies and hashes credentials.
type TokenService struct {
	jwtSecret []byte
	ttl       time.Duration
}

// NewTokenService returns a TokenService bound to a JWT signing secret.
func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{jwtSecret: []byte(secret), ttl: ttl}
}

// Claims are the JWT claims issued for interactive sessions.
type Claims struct {
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

// IssueSessionToken mints a signed JWT for an authenticated user.
func (t *TokenService) IssueSessionToken(ownerID, name, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		OwnerID: ownerID,
		Name:    name,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "custodian",
			Subject:   ownerID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.jwtSecret)
}

// VerifySessionToken validates a JWT and returns its claims.
func (t *TokenService) VerifySessionToken(raw string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return t.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// NewAPIToken generates a random API token with a human-readable prefix.
// The returned plaintext is shown exactly once; only its hash is stored.
func NewAPIToken(name string) (prefix, plaintext, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", "", err
	}
	plaintext = base64.RawURLEncoding.EncodeToString(raw)
	prefix = fmt.Sprintf("cst_%s", strings.ToLower(name))
	if len(prefix) > 20 {
		prefix = prefix[:20]
	}
	hash = HashToken(prefix + "." + plaintext)
	return prefix, plaintext, hash, nil
}

// HashToken returns the SHA-256 hex digest of a token string.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
