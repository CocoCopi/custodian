// Package models defines the core domain entities persisted by the
// Custodian control plane.
package models

import "time"

// ServiceStatus reflects the lifecycle of an application service.
type ServiceStatus string

const (
	StatusProvisioning ServiceStatus = "provisioning"
	StatusRunning      ServiceStatus = "running"
	StatusDeploying    ServiceStatus = "deploying"
	StatusDegraded     ServiceStatus = "degraded"
	StatusStopped      ServiceStatus = "stopped"
	StatusFailed       ServiceStatus = "failed"
)

// Service is an application (or add-on) deployed on the platform.
type Service struct {
	ID        string        `json:"id"`
	OwnerID   string        `json:"owner_id"`
	Name      string        `json:"name"`
	RepoURL   string        `json:"repo_url,omitempty"`
	Branch    string        `json:"branch"`
	BuildType string        `json:"build_type"` // "dockerfile" | "buildpacks" | "static"
	Image     string        `json:"image,omitempty"`
	Blueprint string        `json:"blueprint,omitempty"` // inline custodian.yaml spec
	Status    ServiceStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// Deployment represents a single deploy attempt for a Service.
type Deployment struct {
	ID        string        `json:"id"`
	ServiceID string        `json:"service_id"`
	CommitSHA string        `json:"commit_sha,omitempty"`
	Status    ServiceStatus `json:"status"` // reuses lifecycle states
	Image     string        `json:"image,omitempty"`
	Logs      string        `json:"logs,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	FinishedAt *time.Time   `json:"finished_at,omitempty"`
}

// APIToken is a long-lived credential used by the CLI and CI systems.
type APIToken struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	TokenHash string    `json:"-"`
	Prefix    string    `json:"prefix"`
	CreatedAt time.Time `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

// LogEntry is a single streaming log line captured from a deploy or runtime.
type LogEntry struct {
	DeploymentID string    `json:"deployment_id"`
	ServiceID    string    `json:"service_id"`
	Stream       string    `json:"stream"` // stdout | stderr
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}
