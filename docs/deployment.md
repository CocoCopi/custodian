# Deployment

Custodian ships two deployment profiles behind one control plane. Start with
`compose`; move to `k3s` when you need horizontal autoscaling, scale-to-zero
or multi-node resilience.

## Profile 1 — Single server (`compose`)

Minimum: a VPS with 2 vCPU / 4 GB RAM, Docker 24+ and a domain pointing at it.

```bash
git clone https://github.com/CocoCopi/custodian.git
cd custodian
cp deploy/.env.example deploy/.env
#   set CUSTODIAN_DOMAIN, CUSTODIAN_JWT_SECRET, CUSTODIAN_LETSENCRYPT_EMAIL
make up
```

This brings up:

- `api` + `worker` (control plane, built from `deploy/Dockerfile.api`)
- `postgres`, `redis` (state)
- `console` (React dashboard, built from `frontend/Dockerfile`)
- `traefik` (edge, automatic Let's Encrypt)
- `minio`, `prometheus`, `grafana`, `loki`, `alertmanager` (platform services)

Endpoints: `https://<domain>` (console), `https://api.<domain>` (API),
`https://grafana.<domain>` (dashboards).

### Environment variables

| Variable | Required | Description |
| --- | --- | --- |
| `CUSTODIAN_DOMAIN` | ✅ | Public domain for the console |
| `CUSTODIAN_JWT_SECRET` | ✅ | Signing secret; `openssl rand -hex 32` |
| `CUSTODIAN_DB_*` | — | Postgres credentials (defaults provided) |
| `CUSTODIAN_ENGINE` | — | `compose` (default) or `k3s` |
| `CUSTODIAN_OIDC_*` | — | OIDC provider settings (optional) |
| `CUSTODIAN_LETSENCRYPT_EMAIL` | — | ACME contact email |

## Profile 2 — Cluster (`k3s`)

Requirements: a k3s cluster (or any Kubernetes), cert-manager, Traefik
(default ingress in k3s), and optionally Longhorn + KEDA.

```bash
# Prereqs
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
helm repo add kedacore https://kedacore.github.io/charts && helm install keda kedacore/keda --namespace keda --create-namespace

# Control plane
kubectl apply -f deploy/k3s/
kubectl create secret generic custodian-secrets --namespace custodian \
  --from-literal=jwt-secret="$(openssl rand -hex 32)"
```

With `CUSTODIAN_ENGINE=k3s`, the worker renders and applies Kubernetes
manifests directly: Deployments, Services, HPAs, Ingresses and PVCs — plus
KEDA `ScaledObject`s when a blueprint enables scale-to-zero.

## Upgrades

```bash
git pull
make up          # compose: recreates changed containers
kubectl rollout restart deployment/custodian-api deployment/custodian-worker -n custodian
```

The database schema is applied idempotently at startup
(`internal/store/migrations`), so control plane upgrades are safe to roll
forward.

## Backups

- **Postgres**: dump the `custodian` database (`pg_dump`) or use the Zalando
  Postgres Operator in k3s mode for automated backups and point-in-time
  recovery.
- **Redis**: AOF persistence is enabled (`appendonly yes`); for production
  use, back up `/data` or run Redis Sentinel / Cluster.
- **Volumes**: Longhorn (k3s) or a volume snapshot of the Docker volumes.
