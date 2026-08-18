# Architecture

Custodian is a **control plane** for self-hosted PaaS workloads. It separates
the *decision* work (the API, database and queue) from the *execution* work
(the build and apply workers), so both can scale independently.

## Components

| Component | Technology | Responsibility |
| --- | --- | --- |
| `custodian-api` | Go (Gin) | REST API, OIDC + token auth, WebSocket log hub, blueprint validation, artifact rendering |
| `custodian-worker` | Go (Asynq) | Consumes build/apply jobs, shells out to Docker / Buildpacks / kubectl |
| PostgreSQL | pgx/v5 | Services, deployments, API tokens (schema in `internal/store/migrations`) |
| Redis + Asynq | go-redis, hibiken/asynq | Durable job queue with retries and timeouts |
| Traefik | v3 | Edge router, automatic Let's Encrypt via cert-manager / ACME resolver |
| k3s (optional) | Kubernetes | Runtime when `CUSTODIAN_ENGINE=k3s`: HPA, KEDA, Longhorn, Ingress |
| MinIO | S3-compatible | Object storage for static assets and build artifacts |
| Prometheus / Grafana / Loki | — | Metrics, dashboards, log aggregation |

## Request flow: `custodian deploy hello`

```
CLI ──POST /api/v1/services/hello/deployments──▶ API
  1. API loads the service + its blueprint, validates it
  2. API renders artifacts to <deploy-root>/<project>/<service>/
     - compose engine → docker-compose.yml
     - k3s engine     → Deployment/Service/HPA/Ingress/PVC manifests
  3. API enqueues  build  (deploy:build)   — docker build / pack build + push
  4. API enqueues  apply  (deploy:apply)   — docker compose up -d / kubectl apply
  5. Worker streams every log line into the WebSocket hub
  6. CLI / console receive live logs via WS; status updates via REST polling
```

## Key design decisions

### Jobs are durable, not in-process

Builds can take minutes. If the API restarts mid-build, the job must survive.
Asynq persists tasks in Redis with retries and timeouts, and the worker is a
separate process (`custodian-worker`) that can be scaled horizontally.

### Blueprints are the single source of truth

A `custodian.yaml` is parsed and validated by `internal/blueprint`, then
rendered by `internal/deployer` into engine-specific artifacts. This means one
spec, two targets (compose / k3s), and no drift between environments: the file
*is* the infrastructure.

### Logs are streamed, not polled

The worker publishes `models.LogEntry` events to the in-process hub
(`internal/ws`), which fans them out to WebSocket subscribers keyed by service
and deployment. The hub drops slow consumers instead of blocking the worker —
deploys never stall because a browser tab is closed.

### Auth has two layers

- **OIDC** (Keycloak / Authelia / Google / any provider) for interactive
  sessions via the authorization-code flow (`internal/auth/oidc.go`).
- **API tokens** (`cst_*`) for CLI and CI. Only a SHA-256 hash is stored; the
  plaintext is shown exactly once at creation.

## Security model

- Secrets never touch the database: `valueFrom: secret` references are injected
  at render time from the platform secret store.
- The API requires a valid credential on every route except `/healthz` and the
  OIDC endpoints.
- Worker container has access to the Docker socket only on the control plane
  host; in k3s mode the worker runs in-cluster with a scoped ServiceAccount.
- TLS everywhere via Traefik + cert-manager (Let's Encrypt).

## Observability

The control plane itself ships with Prometheus, Grafana, Loki and
Alertmanager (see `deploy/observability/`). Grafana datasources are
provisioned automatically. OpenTelemetry export hooks (`CUSTODIAN_OTLP_ENDPOINT`)
are available for tracing.

See [deployment.md](deployment.md) for environment variables and
[api.md](api.md) for the full REST surface.
