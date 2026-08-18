package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/CocoCopi/custodian/internal/models"
	"github.com/hibiken/asynq"
)

// Logger receives streaming log lines produced while a job runs.
type Logger interface {
	Log(serviceID, deploymentID, stream, message string)
}

// Store is the subset of the data layer the worker needs.
type Store interface {
	UpdateDeployment(ctx context.Context, id string, status models.ServiceStatus, logs string) error
	UpdateServiceStatus(ctx context.Context, id string, status models.ServiceStatus) error
}

// Worker executes queued build/apply tasks.
type Worker struct {
	store    Store
	logs     Logger
	registry string
	engine   string
	deployRoot string
}

// NewWorker builds the task handler.
func NewWorker(store Store, logs Logger, registry, engine, deployRoot string) *Worker {
	return &Worker{store: store, logs: logs, registry: registry, engine: engine, deployRoot: deployRoot}
}

// Handler returns the Asynq mux registering all task types.
func (w *Worker) Handler() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeBuild, w.handleBuild)
	mux.HandleFunc(TypeApply, w.handleApply)
	return mux
}

func (w *Worker) handleBuild(ctx context.Context, t *asynq.Task) error {
	var p BuildPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("decode build payload: %w", err)
	}
	w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", "building image for "+p.ServiceName+"@"+p.Branch)

	image := fmt.Sprintf("%s/custodian/%s/%s:%s", w.registry, p.ProjectName, p.ServiceName, shortSHA(p.CommitSHA))
	workDir := filepath.Join(w.deployRoot, p.ProjectName, p.ServiceName)

	switch p.BuildType {
	case "dockerfile":
		cmd := exec.CommandContext(ctx, "docker", "build",
			"--tag", image,
			"--file", filepath.Join(workDir, "Dockerfile"),
			workDir)
		out, err := cmd.CombinedOutput()
		w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", string(out))
		if err != nil {
			_ = w.store.UpdateDeployment(ctx, p.DeploymentID, models.StatusFailed, string(out))
			return fmt.Errorf("docker build: %w", err)
		}
	case "buildpacks":
		cmd := exec.CommandContext(ctx, "pack", "build", image,
			"--builder", "paketobuildpacks/builder-jammy-base",
			"--path", workDir)
		out, err := cmd.CombinedOutput()
		w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", string(out))
		if err != nil {
			_ = w.store.UpdateDeployment(ctx, p.DeploymentID, models.StatusFailed, string(out))
			return fmt.Errorf("pack build: %w", err)
		}
	case "static":
		// Static sites are served from the built artifact; image step is a no-op.
		w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", "static site: skipping image build")
	}

	push := exec.CommandContext(ctx, "docker", "push", image)
	if out, err := push.CombinedOutput(); err != nil {
		w.logs.Log(p.ServiceID, p.DeploymentID, "stderr", string(out))
		return fmt.Errorf("docker push: %w", err)
	}
	w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", "image "+image+" pushed")
	return nil
}

func (w *Worker) handleApply(ctx context.Context, t *asynq.Task) error {
	var p ApplyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("decode apply payload: %w", err)
	}
	w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", "applying "+p.ServiceName+" ("+p.Engine+")")

	var cmd *exec.Cmd
	switch p.Engine {
	case "compose":
		cmd = exec.CommandContext(ctx, "docker", "compose", "-f", filepath.Join(p.ArtifactsDir, "docker-compose.yml"), "up", "-d", "--no-deps", p.ServiceName)
	case "k3s":
		cmd = exec.CommandContext(ctx, "kubectl", "apply", "-f", p.ArtifactsDir)
	default:
		return fmt.Errorf("unsupported engine %q", p.Engine)
	}
	out, err := cmd.CombinedOutput()
	w.logs.Log(p.ServiceID, p.DeploymentID, "stdout", string(out))
	if err != nil {
		_ = w.store.UpdateDeployment(ctx, p.DeploymentID, models.StatusFailed, string(out))
		return fmt.Errorf("apply: %w", err)
	}
	if err := w.store.UpdateDeployment(ctx, p.DeploymentID, models.StatusRunning, string(out)); err != nil {
		return err
	}
	return w.store.UpdateServiceStatus(ctx, p.ServiceID, models.StatusRunning)
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	if sha == "" {
		return "latest"
	}
	return sha
}
