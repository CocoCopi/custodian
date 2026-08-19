package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/CocoCopi/custodian/internal/auth"
	"github.com/CocoCopi/custodian/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// handleSetupStatus checks if any user accounts exist in the control plane.
func (s *Server) handleSetupStatus(c *gin.Context) {
	count, err := s.store.CountUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"setup_required": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"setup_required": count == 0})
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

// handleRegister provisions a local user account.
func (s *Server) handleRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 6 characters long"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := s.store.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists or invalid"})
		return
	}

	jwtToken, err := s.tokens.IssueSessionToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue session token"})
		return
	}

	c.SetCookie("custodian_session", jwtToken, int(s.cfg.TokenTTL.Seconds()), "/", "", false, true)
	c.JSON(http.StatusCreated, gin.H{"token": jwtToken, "user": user.Username})
}

type localLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// handleLocalLogin authenticates local user account credentials.
func (s *Server) handleLocalLogin(c *gin.Context) {
	var req localLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	// 1. Check registered database users
	if user, err := s.store.GetUserByUsername(c.Request.Context(), req.Username); err == nil && user != nil {
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) == nil {
			jwtToken, err := s.tokens.IssueSessionToken(user.ID, user.Username, user.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue session token"})
				return
			}
			c.SetCookie("custodian_session", jwtToken, int(s.cfg.TokenTTL.Seconds()), "/", "", false, true)
			c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": user.Username})
			return
		}
	}

	// 2. Fallback check for config admin user
	if req.Username == s.cfg.AdminUser && req.Password == s.cfg.AdminPassword {
		jwtToken, err := s.tokens.IssueSessionToken(req.Username, req.Username, req.Username+"@localhost")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue session token"})
			return
		}
		c.SetCookie("custodian_session", jwtToken, int(s.cfg.TokenTTL.Seconds()), "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": req.Username})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
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
		"token":     t,
		"plaintext": prefix + "." + plaintext,
		"note":      "store this value now; it will not be shown again",
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
