// Package jobs orchestrates asynchronous build and deploy work using Asynq
// on top of Redis.
package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

// Task types registered with the Asynq server.
const (
	TypeBuild   = "deploy:build"   // build image from repo
	TypeApply   = "deploy:apply"   // apply rendered artifacts
	TypeFinalize = "deploy:finalize" // mark deployment complete
)

// BuildPayload describes a build job.
type BuildPayload struct {
	DeploymentID string `json:"deployment_id"`
	ServiceID    string `json:"service_id"`
	ProjectName  string `json:"project_name"`
	ServiceName  string `json:"service_name"`
	RepoURL      string `json:"repo_url"`
	Branch       string `json:"branch"`
	CommitSHA    string `json:"commit_sha"`
	BuildType    string `json:"build_type"`
	BuildContext string `json:"build_context"`
	Registry     string `json:"registry"`
}

// ApplyPayload describes an artifact apply job.
type ApplyPayload struct {
	DeploymentID string `json:"deployment_id"`
	ServiceID    string `json:"service_id"`
	Engine       string `json:"engine"` // compose | k3s
	ProjectName  string `json:"project_name"`
	ServiceName  string `json:"service_name"`
	ArtifactsDir string `json:"artifacts_dir"`
}

// Client enqueues tasks.
type Client struct {
	inner *asynq.Client
}

// NewClient connects the task producer to Redis.
func NewClient(redisAddr, redisPassword string) *Client {
	return &Client{inner: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword})}
}

// Close releases the client connection.
func (c *Client) Close() error { return c.inner.Close() }

// EnqueueBuild schedules a build task with retries.
func (c *Client) EnqueueBuild(ctx context.Context, p BuildPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = c.inner.EnqueueContext(ctx, asynq.NewTask(TypeBuild, payload), asynq.MaxRetry(5), asynq.Timeout(30*time.Minute))
	return err
}

// EnqueueApply schedules an artifact apply task.
func (c *Client) EnqueueApply(ctx context.Context, p ApplyPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = c.inner.EnqueueContext(ctx, asynq.NewTask(TypeApply, payload), asynq.MaxRetry(3), asynq.Timeout(15*time.Minute))
	return err
}
