package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys used to pass the authenticated identity to handlers.
const (
	CtxOwnerID = "owner_id"
	CtxSubject = "subject"
)

// Middleware authenticates requests via either a Bearer session JWT or a
// Custodian API token (Authorization: Bearer cst_<name>.<secret>).
type Middleware struct {
	tokens   *TokenService
	resolver func(ctx *gin.Context, hash string) (ownerID, tokenID string, ok bool)
}

// NewMiddleware builds the auth middleware. resolver looks up an API token
// hash and returns the owning account; it may be nil to disable token auth.
func NewMiddleware(tokens *TokenService, resolver func(ctx *gin.Context, hash string) (string, string, bool)) *Middleware {
	return &Middleware{tokens: tokens, resolver: resolver}
}

// RequireAuth enforces that the request carries a valid credential.
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization scheme"})
			return
		}
		token := strings.TrimSpace(parts[1])

		// Session JWT first.
		if claims, err := m.tokens.VerifySessionToken(token); err == nil {
			c.Set(CtxOwnerID, claims.OwnerID)
			c.Set(CtxSubject, claims.Subject)
			c.Next()
			return
		}

		// Fall back to API token lookup.
		if m.resolver != nil && strings.HasPrefix(token, "cst_") {
			hash := HashToken(token)
			if ownerID, _, ok := m.resolver(c, hash); ok {
				c.Set(CtxOwnerID, ownerID)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired credentials"})
	}
}

// OptionalAuth attaches identity when present but never rejects the request.
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if claims, err := m.tokens.VerifySessionToken(token); err == nil {
				c.Set(CtxOwnerID, claims.OwnerID)
				c.Set(CtxSubject, claims.Subject)
			}
		}
		c.Next()
	}
}
