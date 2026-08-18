package api

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/CocoCopi/custodian/internal/blueprint"
	"github.com/CocoCopi/custodian/internal/deployer"
	"github.com/CocoCopi/custodian/internal/jobs"
	"github.com/CocoCopi/custodian/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true }, // tightened behind auth + CORS in production
}

// handleCreateDeployment triggers a full build + apply pipeline for a service.
func (s *Server) handleCreateDeployment(c *gin.Context) {
	serviceID := c.Param("id")
	svc, err := s.store.GetService(c.Request.Context(), serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	// Resolve the blueprint: inline spec wins, else a bare image deploy.
	project := svc.Name
	var svcSpec *blueprint.Service
	if svc.Blueprint != "" {
		bp, err := blueprint.Parse([]byte(svc.Blueprint))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid blueprint: " + err.Error()})
			return
		}
		project = bp.Name
		if s, ok := bp.FindService(svc.Name); ok {
			svcSpec = s
		} else if len(bp.Services) == 1 {
			svcSpec = &bp.Services[0]
		}
	}
	if svcSpec == nil && svc.Image == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service has no image and no usable blueprint"})
		return
	}

	deployment := &models.Deployment{
		ServiceID: serviceID,
		CommitSHA: c.Query("commit"),
		Status:    models.StatusProvisioning,
	}
	if err := s.store.CreateDeployment(c.Request.Context(), deployment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Render artifacts to disk so the worker can apply them.
	artifactsDir, err := s.writeArtifacts(project, svc, svcSpec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "render artifacts: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	if svcSpec != nil {
		buildType := svcSpec.Build.Type
		if buildType == "" {
			buildType = svc.BuildType
		}
		if err := s.queue.EnqueueBuild(ctx, jobs.BuildPayload{
			DeploymentID: deployment.ID,
			ServiceID:    serviceID,
			ProjectName:  project,
			ServiceName:  svc.Name,
			RepoURL:      svc.RepoURL,
			Branch:       svc.Branch,
			CommitSHA:    deployment.CommitSHA,
			BuildType:    buildType,
			BuildContext: svcSpec.Build.Context,
			Registry:     s.registry,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "enqueue build: " + err.Error()})
			return
		}
	}
	if err := s.queue.EnqueueApply(ctx, jobs.ApplyPayload{
		DeploymentID: deployment.ID,
		ServiceID:    serviceID,
		Engine:       string(s.engine),
		ProjectName:  project,
		ServiceName:  svc.Name,
		ArtifactsDir: artifactsDir,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "enqueue apply: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, deployment)
}

// writeArtifacts renders the blueprint into deployable files on disk.
func (s *Server) writeArtifacts(project string, svc *models.Service, spec *blueprint.Service) (string, error) {
	dir := filepath.Join(s.deployRoot, project, svc.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	if spec == nil {
		// Bare image deploys still get a compose manifest.
		bp := &blueprint.Blueprint{Name: project, Services: []blueprint.Service{{
			Name: svc.Name,
			Build: blueprint.Build{Type: "dockerfile"},
			Runtime: blueprint.Runtime{Replicas: 1, Port: 8080},
		}}}
		spec = &bp.Services[0]
	}

	result, err := deployer.Render(s.engine, &blueprint.Blueprint{Name: project, Services: []blueprint.Service{*spec}}, s.registry)
	if err != nil {
		return "", err
	}

	switch s.engine {
	case deployer.EngineCompose:
		if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), result.Compose, 0o644); err != nil {
			return "", err
		}
	case deployer.EngineK3s:
		for name, content := range result.Manifests {
			if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
				return "", err
			}
		}
	}
	return dir, nil
}

// handleListDeployments returns the deploy history for a service.
func (s *Server) handleListDeployments(c *gin.Context) {
	deployments, err := s.store.ListDeployments(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deployments": deployments})
}

// handleGetDeployment returns a single deployment with its log tail.
func (s *Server) handleGetDeployment(c *gin.Context) {
	d, err := s.store.GetDeployment(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

// handleDeploymentLogs streams live logs for a deployment over WebSocket.
func (s *Server) handleDeploymentLogs(c *gin.Context) {
	s.streamLogs(c, "", c.Param("id"))
}

// handleServiceLogs streams live logs for a service over WebSocket.
func (s *Server) handleServiceLogs(c *gin.Context) {
	s.streamLogs(c, c.Param("id"), "")
}

func (s *Server) streamLogs(c *gin.Context, serviceID, deploymentID string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	topic := serviceID
	if topic == "" {
		topic = deploymentID
	}
	ch, unsubscribe := s.hub.Subscribe(topic)
	defer unsubscribe()

	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(nowPlus60s()) })
	_ = conn.SetReadDeadline(nowPlus60s())
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for entry := range ch {
		if err := conn.WriteJSON(entry); err != nil {
			return
		}
	}
}

func nowPlus60s() time.Time {
	return time.Now().Add(60 * time.Second)
}
