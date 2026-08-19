// Package store provides PostgreSQL persistence for the control plane.
package store

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/CocoCopi/custodian/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps the connection pool and exposes typed repositories.
type Store struct {
	pool *pgxpool.Pool
}

// New connects to PostgreSQL and applies the embedded schema.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		sqlBytes, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", e.Name(), err)
		}
	}
	return nil
}

// Close releases the connection pool.
func (s *Store) Close() { s.pool.Close() }

// ---- Services ----

// CreateService inserts a new service.
func (s *Store) CreateService(ctx context.Context, svc *models.Service) error {
	svc.ID = uuid.NewString()
	svc.CreatedAt = time.Now().UTC()
	svc.UpdatedAt = svc.CreatedAt
	_, err := s.pool.Exec(ctx, `
		INSERT INTO services (id, owner_id, name, repo_url, branch, build_type, image, blueprint, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		svc.ID, svc.OwnerID, svc.Name, svc.RepoURL, svc.Branch, svc.BuildType, svc.Image, svc.Blueprint, svc.Status, svc.CreatedAt, svc.UpdatedAt)
	return err
}

// GetService returns a service by id.
func (s *Store) GetService(ctx context.Context, id string) (*models.Service, error) {
	var svc models.Service
	err := s.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, repo_url, branch, build_type, image, blueprint, status, created_at, updated_at
		FROM services WHERE id = $1`, id).
		Scan(&svc.ID, &svc.OwnerID, &svc.Name, &svc.RepoURL, &svc.Branch, &svc.BuildType, &svc.Image, &svc.Blueprint, &svc.Status, &svc.CreatedAt, &svc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// ListServices returns all services for an owner, newest first.
func (s *Store) ListServices(ctx context.Context, ownerID string) ([]models.Service, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, owner_id, name, repo_url, branch, build_type, image, blueprint, status, created_at, updated_at
		FROM services WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []models.Service{}
	for rows.Next() {
		var svc models.Service
		if err := rows.Scan(&svc.ID, &svc.OwnerID, &svc.Name, &svc.RepoURL, &svc.Branch, &svc.BuildType, &svc.Image, &svc.Blueprint, &svc.Status, &svc.CreatedAt, &svc.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, rows.Err()
}

// UpdateServiceStatus flips a service lifecycle state.
func (s *Store) UpdateServiceStatus(ctx context.Context, id string, status models.ServiceStatus) error {
	_, err := s.pool.Exec(ctx, `UPDATE services SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	return err
}

// UpdateService updates editable service attributes.
func (s *Store) UpdateService(ctx context.Context, svc *models.Service) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE services SET repo_url = $2, branch = $3, build_type = $4, image = $5, blueprint = $6, updated_at = now()
		WHERE id = $1`, svc.ID, svc.RepoURL, svc.Branch, svc.BuildType, svc.Image, svc.Blueprint)
	return err
}

// DeleteService removes a service and cascades to deployments.
func (s *Store) DeleteService(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM services WHERE id = $1`, id)
	return err
}

// ---- Deployments ----

// CreateDeployment records a new deploy attempt.
func (s *Store) CreateDeployment(ctx context.Context, d *models.Deployment) error {
	d.ID = uuid.NewString()
	d.CreatedAt = time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO deployments (id, service_id, commit_sha, status, image, logs, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		d.ID, d.ServiceID, d.CommitSHA, d.Status, d.Image, d.Logs, d.CreatedAt)
	return err
}

// GetDeployment returns a deployment by id.
func (s *Store) GetDeployment(ctx context.Context, id string) (*models.Deployment, error) {
	var d models.Deployment
	err := s.pool.QueryRow(ctx, `
		SELECT id, service_id, commit_sha, status, image, logs, created_at, finished_at
		FROM deployments WHERE id = $1`, id).
		Scan(&d.ID, &d.ServiceID, &d.CommitSHA, &d.Status, &d.Image, &d.Logs, &d.CreatedAt, &d.FinishedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDeployments returns deployments for a service, newest first.
func (s *Store) ListDeployments(ctx context.Context, serviceID string) ([]models.Deployment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, service_id, commit_sha, status, image, logs, created_at, finished_at
		FROM deployments WHERE service_id = $1 ORDER BY created_at DESC LIMIT 100`, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	deployments := []models.Deployment{}
	for rows.Next() {
		var d models.Deployment
		if err := rows.Scan(&d.ID, &d.ServiceID, &d.CommitSHA, &d.Status, &d.Image, &d.Logs, &d.CreatedAt, &d.FinishedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}
	return deployments, rows.Err()
}

// UpdateDeployment writes status and/or logs for a deployment.
func (s *Store) UpdateDeployment(ctx context.Context, id string, status models.ServiceStatus, logs string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE deployments SET status = $2, logs = $3, finished_at = CASE WHEN $2 IN ('running','failed','stopped') THEN now() ELSE finished_at END
		WHERE id = $1`, id, status, logs)
	return err
}

// ---- API tokens ----

// CreateAPIToken persists a token record.
func (s *Store) CreateAPIToken(ctx context.Context, t *models.APIToken) error {
	t.ID = uuid.NewString()
	t.CreatedAt = time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO api_tokens (id, name, owner_id, token_hash, prefix, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.Name, t.OwnerID, t.TokenHash, t.Prefix, t.CreatedAt)
	return err
}

// GetAPITokenByHash returns a token record matching a stored hash.
func (s *Store) GetAPITokenByHash(ctx context.Context, hash string) (*models.APIToken, error) {
	var t models.APIToken
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, owner_id, token_hash, prefix, created_at, last_used_at
		FROM api_tokens WHERE token_hash = $1`, hash).
		Scan(&t.ID, &t.Name, &t.OwnerID, &t.TokenHash, &t.Prefix, &t.CreatedAt, &t.LastUsedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListAPITokens returns tokens for an owner.
func (s *Store) ListAPITokens(ctx context.Context, ownerID string) ([]models.APIToken, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, owner_id, token_hash, prefix, created_at, last_used_at
		FROM api_tokens WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := []models.APIToken{}
	for rows.Next() {
		var t models.APIToken
		if err := rows.Scan(&t.ID, &t.Name, &t.OwnerID, &t.TokenHash, &t.Prefix, &t.CreatedAt, &t.LastUsedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// DeleteAPIToken revokes a token.
func (s *Store) DeleteAPIToken(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM api_tokens WHERE id = $1`, id)
	return err
}

// TouchAPIToken bumps last_used_at.
func (s *Store) TouchAPIToken(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE api_tokens SET last_used_at = now() WHERE id = $1`, id)
	return err
}

// ---- Users ----

// CountUsers returns the total count of registered users.
func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&count)
	return count, err
}

// CreateUser inserts a new user record.
func (s *Store) CreateUser(ctx context.Context, u *models.User) error {
	u.ID = uuid.NewString()
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.Username, u.PasswordHash, u.Email, u.CreatedAt, u.UpdatedAt)
	return err
}

// GetUserByUsername fetches a user by their username.
func (s *Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, password_hash, email, created_at, updated_at
		FROM users WHERE LOWER(username) = LOWER($1)`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
