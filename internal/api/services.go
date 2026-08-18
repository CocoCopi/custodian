package api

import (
	"net/http"
	"strings"

	"github.com/CocoCopi/custodian/internal/auth"
	"github.com/CocoCopi/custodian/internal/blueprint"
	"github.com/CocoCopi/custodian/internal/models"
	"github.com/gin-gonic/gin"
)

type createServiceRequest struct {
	Name      string `json:"name" binding:"required"`
	RepoURL   string `json:"repo_url"`
	Branch    string `json:"branch"`
	BuildType string `json:"build_type"`
	Image     string `json:"image"`
	Blueprint string `json:"blueprint"` // inline custodian.yaml
}

// handleCreateService provisions a new service from an inline blueprint.
func (s *Server) handleCreateService(c *gin.Context) {
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	// Validate the blueprint when provided so bad specs fail fast.
	if req.Blueprint != "" {
		if _, err := blueprint.Parse([]byte(req.Blueprint)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid blueprint: " + err.Error()})
			return
		}
	}

	buildType := req.BuildType
	if buildType == "" {
		buildType = "dockerfile"
	}
	svc := &models.Service{
		OwnerID:   c.GetString(auth.CtxOwnerID),
		Name:      strings.ToLower(req.Name),
		RepoURL:   req.RepoURL,
		Branch:    req.Branch,
		BuildType: buildType,
		Image:     req.Image,
		Blueprint: req.Blueprint,
		Status:    models.StatusProvisioning,
	}
	if svc.Branch == "" {
		svc.Branch = "main"
	}
	if err := s.store.CreateService(c.Request.Context(), svc); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "a service with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, svc)
}

// handleListServices returns the caller's services.
func (s *Server) handleListServices(c *gin.Context) {
	ownerID := c.GetString(auth.CtxOwnerID)
	services, err := s.store.ListServices(c.Request.Context(), ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"services": services})
}

// handleGetService returns a single service.
func (s *Server) handleGetService(c *gin.Context) {
	svc, err := s.store.GetService(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	c.JSON(http.StatusOK, svc)
}

// handleDeleteService removes a service and its deployments.
func (s *Server) handleDeleteService(c *gin.Context) {
	if err := s.store.DeleteService(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
