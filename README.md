<div align="center">

# 🛡️ Custodian

**The self-hosted PaaS that gives you full ownership.**

Render/Heroku-style deployment for your own infrastructure — no vendor lock-in,
no per-seat fees, no black boxes. Deploy apps, databases and static sites with a
single blueprint file, and keep every byte of your platform on servers you control.

[![CI](https://github.com/CocoCopi/custodian/actions/workflows/ci.yml/badge.svg)](https://github.com/CocoCopi/custodian/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/CocoCopi/custodian)](https://goreportcard.com/report/github.com/CocoCopi/custodian)
[![License: Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23-blue)](go.mod)

**Docs · [Architecture](docs/architecture.md) · [Blueprints](docs/blueprint.md) · [API](docs/api.md) · [Deployment](docs/deployment.md)**

</div>

---

## Why Custodian?

Platform tools like [Render](https://render.com) and [Heroku](https://heroku.com)
are wonderful — until you need full ownership. Self-hosted alternatives like
Coolify, CapRover and Dokku give you the server, but leave **you** as the ops
team: patching kernels, babysitting disks, hand-rolling backups and fighting
with DNS.

Custodian closes that gap. It is a **control plane** for your own fleet:

| Capability | Custodian | Coolify / CapRover / Dokku |
| --- | --- | --- |
| Deploy from `git push` / GitHub & GitLab webhooks | ✅ | partial |
| Declarative infra-as-code (`custodian.yaml`) | ✅ | partial / none |
| Horizontal autoscaling (CPU / memory / queue) | ✅ HPA + KEDA | single-server only |
| Scale-to-zero | ✅ KEDA | ❌ |
| Live logs over WebSocket (console + CLI) | ✅ | varies |
| Managed TLS via cert-manager + Let's Encrypt | ✅ | partial |
| Object storage (S3) via MinIO | ✅ | ❌ |
| Full observability (Prometheus, Grafana, Loki) | ✅ | partial |
| OIDC SSO + API tokens for CI | ✅ | minimal |
| Zero-downtime deploys with health-check rollbacks | ✅ (blueprint probes) | needs tuning |
| Global edge / CDN | Cloudflare in front | ❌ |

**Custodian does not hide the server from you — it puts you in charge of it.**
Own your data, own your uptime, own your platform.

---

## Architecture

```
                        ┌────────────────────────────────────────────┐
   Browser / CLI  ───▶  │                Control Plane               │
                        │                                            │
                        │   Go API (Gin) ──── WebSocket log hub      │
                        │        │                 │                 │
                        │   PostgreSQL ◀─── Redis / Asynq queue      │
                        │        │                 │                 │
                        └────────┼─────────────────┼─────────────────┘
                                 │                 │  build/deploy jobs
                                 ▼                 ▼
                        ┌────────────────────────────────────────────┐
                        │               Build System                 │
                        │   Docker / BuildKit  ·  Buildpacks (pack)  │
                        │   Container Registry (Harbor / local)      │
                        └────────────────────────────────────────────┘
                                 │
                                 ▼
                        ┌────────────────────────────────────────────┐
                        │        Runtime / Orchestration             │
                        │   k3s (Kubernetes)  ·  Docker Compose      │
                        │   Traefik  ·  cert-manager  ·  CoreDNS     │
                        │   Longhorn (PV)  ·  MinIO (S3)             │
                        └────────────────────────────────────────────┘
                                 │
                                 ▼
                        ┌────────────────────────────────────────────┐
                        │              Observability                 │
                        │   Prometheus · Grafana · Loki · Alertmanager│
                        └────────────────────────────────────────────┘
```

Two deployment profiles, one control plane:

- **`compose` (default)** — single server: Docker + Traefik. Perfect for a
  VPS, a homelab, or a small team. Everything the control plane needs runs via
  `docker compose`.
- **`k3s`** — lightweight Kubernetes cluster: HPA autoscaling, KEDA
  scale-to-zero, Longhorn persistent volumes, cert-manager TLS.

## Quickstart

### 1. Start the control plane

```bash
git clone https://github.com/CocoCopi/custodian.git
cd custodian
cp deploy/.env.example deploy/.env   # set CUSTODIAN_JWT_SECRET and a domain
make up                              # docker compose -f deploy/docker-compose.yml up -d
```

The console is now at `https://<your-domain>`, the API at
`https://api.<your-domain>`, and Grafana at `https://grafana.<your-domain>`.

### 2. Create a token and log in

```bash
# From the control plane host (bootstraps the first token via the API):
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Authorization: Bearer <bootstrap-token>" \
  -d '{"name":"cli"}'

# Or simply use the CLI against a local instance:
go build -o bin/custodian ./cmd/custodian-cli
./bin/custodian login
```

### 3. Deploy your first app

```yaml
# custodian.yaml
apiVersion: custodian.dev/v1
kind: Blueprint
name: hello
services:
  - name: web
    build:
      type: dockerfile
    runtime:
      port: 8080
      env:
        - name: MESSAGE
          value: hello from custodian
    healthCheck:
      path: /healthz
    autoscaling:
      enabled: true
      minReplicas: 1
      maxReplicas: 5
    domains:
      - hello.example.com
```

```bash
./bin/custodian apps create hello --blueprint custodian.yaml
./bin/custodian deploy hello
./bin/custodian logs hello --follow
```

That's it — Custodian builds the image, applies the manifests, wires up
Traefik routing and a Let's Encrypt certificate, and starts streaming logs.

## CLI

```
custodian login                 Authenticate with the control plane
custodian apps create <name>    Create an application (--blueprint, --repo)
custodian apps list             List applications
custodian apps get <name>       Show application details
custodian apps delete <name>    Remove an application
custodian deploy <name>         Trigger a deployment (--commit)
custodian logs <name>           Stream live logs (--follow)
custodian tokens create <name>  Issue an API token for CI/CD
```

## The blueprint

A `custodian.yaml` describes an entire project — services, build, environment,
autoscaling, domains and persistence — as code. It is the single source of
truth for a deployment, and the same file works on a single VPS or a k3s
cluster. See [docs/blueprint.md](docs/blueprint.md) for the full reference and
[examples/custodian.yaml](examples/custodian.yaml) for a complete example.

## Roadmap

- [ ] GitHub / GitLab webhook triggers (auto-deploy on push)
- [ ] Preview environments per pull request
- [ ] Managed Postgres/Redis with automated backups and point-in-time recovery
- [ ] Zero-downtime deploys with health-check rollbacks
- [ ] Audit log + role-based access control for teams
- [ ] Harbor registry integration and multi-region deploys
- [ ] Terraform provider + `custodian` provider for IaC workflows

## Development

```bash
make test          # backend tests
make vet           # go vet
make lint          # golangci-lint
make frontend-build # build the React console
```

Backend: Go 1.23+ (Gin, pgx, Asynq, go-oidc) · Frontend: React 18 + TypeScript +
Tailwind + TanStack Query + Zustand. See [CONTRIBUTING.md](CONTRIBUTING.md) to
get involved.

## License

Custodian is licensed under the [Apache License 2.0](LICENSE). Contributions
are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) and our
[Code of Conduct](CODE_OF_CONDUCT.md).
