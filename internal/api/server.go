// Package api exposes the Custodian control plane REST + WebSocket API.
package api

import (
	"net/http"

	"github.com/CocoCopi/custodian/internal/auth"
	"github.com/CocoCopi/custodian/internal/config"
	"github.com/CocoCopi/custodian/internal/deployer"
	"github.com/CocoCopi/custodian/internal/jobs"
	"github.com/CocoCopi/custodian/internal/store"
	"github.com/CocoCopi/custodian/internal/ws"
	"github.com/gin-gonic/gin"
)

// Server wires together all dependencies and builds the HTTP router.
type Server struct {
	cfg        *config.Config
	store      *store.Store
	tokens     *auth.TokenService
	oidc       *auth.OIDCProvider
	queue      *jobs.Client
	hub        *ws.Hub
	engine     deployer.Engine
	registry   string
	deployRoot string
}

// New assembles a Server with all runtime dependencies.
func New(cfg *config.Config, st *store.Store, tokens *auth.TokenService, oidc *auth.OIDCProvider, queue *jobs.Client, hub *ws.Hub) *Server {
	return &Server{
		cfg:        cfg,
		store:      st,
		tokens:     tokens,
		oidc:       oidc,
		queue:      queue,
		hub:        hub,
		engine:     deployer.Engine(cfg.Engine),
		registry:   cfg.PublicURL,
		deployRoot: cfg.DeployRoot,
	}
}

// Router builds the Gin engine with routes and middleware registered.
func (s *Server) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), requestLogger())

	mw := auth.NewMiddleware(s.tokens, func(c *gin.Context, hash string) (string, string, bool) {
		t, err := s.store.GetAPITokenByHash(c.Request.Context(), hash)
		if err != nil {
			return "", "", false
		}
		_ = s.store.TouchAPIToken(c.Request.Context(), t.ID)
		return t.OwnerID, t.ID, true
	})

	r.GET("/healthz", s.handleHealth)
	r.GET("/api/v1/auth/login", s.handleLogin)
	r.GET("/api/v1/auth/callback", s.handleCallback)

	v1 := r.Group("/api/v1")
	v1.Use(mw.RequireAuth())
	{
		v1.GET("/me", s.handleMe)

		v1.GET("/services", s.handleListServices)
		v1.POST("/services", s.handleCreateService)
		v1.GET("/services/:id", s.handleGetService)
		v1.DELETE("/services/:id", s.handleDeleteService)
		v1.GET("/services/:id/deployments", s.handleListDeployments)
		v1.POST("/services/:id/deployments", s.handleCreateDeployment)
		v1.GET("/deployments/:id", s.handleGetDeployment)
		v1.GET("/deployments/:id/logs", s.handleDeploymentLogs)
		v1.GET("/services/:id/logs", s.handleServiceLogs)

		v1.GET("/tokens", s.handleListTokens)
		v1.POST("/tokens", s.handleCreateToken)
		v1.DELETE("/tokens/:id", s.handleDeleteToken)
	}

	return r
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// Logged by the platform's observability stack in production.
		_ = http.StatusOK
	}
}
