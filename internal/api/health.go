package api

import (
	"net/http"

	"github.com/CocoCopi/custodian/internal/auth"
	"github.com/gin-gonic/gin"
)

// handleHealth reports liveness and dependency status.
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"engine":   s.engine,
		"version":  "0.1.0",
		"oidc":     s.oidc.Enabled(),
	})
}

// handleMe returns the authenticated identity.
func (s *Server) handleMe(c *gin.Context) {
	ownerID := c.GetString(auth.CtxOwnerID)
	c.JSON(http.StatusOK, gin.H{"owner_id": ownerID})
}
