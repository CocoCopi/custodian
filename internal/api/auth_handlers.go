package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/CocoCopi/custodian/internal/auth"
	"github.com/CocoCopi/custodian/internal/models"
	"github.com/gin-gonic/gin"
)

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// handleLogin redirects to the OIDC provider to start the auth code flow.
func (s *Server) handleLogin(c *gin.Context) {
	if !s.oidc.Enabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC is not configured on this instance"})
		return
	}
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate state"})
		return
	}
	c.SetCookie("custodian_oauth_state", state, 600, "/", "", false, true)
	c.Redirect(http.StatusFound, s.oidc.AuthCodeURL(state))
}

// handleCallback completes the OIDC exchange and issues a session JWT.
func (s *Server) handleCallback(c *gin.Context) {
	if !s.oidc.Enabled() {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OIDC is not configured on this instance"})
		return
	}
	state, err := c.Cookie("custodian_oauth_state")
	if err != nil || state == "" || state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OAuth state"})
		return
	}
	subject, name, email, err := s.oidc.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	jwtToken, err := s.tokens.IssueSessionToken(subject, name, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue session"})
		return
	}
	c.SetCookie("custodian_session", jwtToken, int(s.cfg.TokenTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

// handleListTokens returns the caller's API tokens (without hashes).
func (s *Server) handleListTokens(c *gin.Context) {
	ownerID := c.GetString(auth.CtxOwnerID)
	tokens, err := s.store.ListAPITokens(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

// handleCreateToken generates a new API token. The plaintext is returned once.
func (s *Server) handleCreateToken(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	ownerID := c.GetString(auth.CtxOwnerID)
	prefix, plaintext, hash, err := auth.NewAPIToken(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	t := &models.APIToken{Name: req.Name, OwnerID: ownerID, TokenHash: hash, Prefix: prefix}
	if err := s.store.CreateAPIToken(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token":      t,
		"plaintext":  prefix + "." + plaintext,
		"note":       "store this value now; it will not be shown again",
	})
}

// handleDeleteToken revokes an API token.
func (s *Server) handleDeleteToken(c *gin.Context) {
	if err := s.store.DeleteAPIToken(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
