# Blueprint reference (`custodian.yaml`)

A blueprint is a declarative description of an entire project. Custodian
parses, validates and renders it into deployable artifacts — the same file
works on a single VPS (`compose` engine) and a k3s cluster (`k3s` engine).

## Top level

```yaml
apiVersion: custodian.dev/v1   # default if omitted
kind: Blueprint                # default if omitted
name: my-project               # required; used for routing and volume names
services: [...]                # required; one or more services
```

## Service

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `name` | string | — | Service name (unique within the blueprint) |
| `build` | object | — | How the image is produced |
| `runtime` | object | — | Environment, resources, replicas |
| `healthCheck` | object | — | Readiness/liveness probe |
| `autoscaling` | object | — | HPA/KEDA scaling policy |
| `domains` | string[] | — | Hostnames routed to this service |
| `persistence` | object[] | — | Persistent volumes |

### `build`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `type` | string | `dockerfile` | `dockerfile`, `buildpacks` or `static` |
| `dockerfile` | string | `Dockerfile` | Path within the build context |
| `context` | string | `.` | Build context directory |
| `buildArgs` | map | — | Build-time arguments (`ARG`) |
| `command` | string | — | Start command (buildpacks) |

### `runtime`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `replicas` | int | `1` | Initial replica count |
| `port` | int | `8080` | Container listen port |
| `env` | object[] | — | Environment variables |
| `resources` | object | — | `cpu` / `memory` limits |

Environment variables:

```yaml
env:
  - name: DATABASE_URL
    value: postgres://...      # literal value
  - name: STRIPE_SECRET
    valueFrom: secret          # injected by the control plane at deploy time
```

### `healthCheck`

```yaml
healthCheck:
  path: /healthz      # required, defaults to "/healthz" if omitted
  interval: 10s       # default 10s
  timeout: 3s         # default 3s
```

Health checks power zero-downtime deploys: a new revision is only routed
traffic once its probe passes, and a failed rollout is rolled back.

### `autoscaling`

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `enabled` | bool | `false` | Enable horizontal autoscaling |
| `minReplicas` | int | `1` | Minimum replicas (`0` enables scale-to-zero via KEDA) |
| `maxReplicas` | int | `10` | Maximum replicas |
| `targetCPU` | int | `70` | CPU utilization target (%) |
| `targetMemory` | int | — | Memory utilization target (%) |
| `scaleToZero` | bool | `false` | Emit a KEDA ScaledObject (k3s engine) |

### `persistence`

```yaml
persistence:
  - name: uploads
    size: 20Gi
    mountPath: /var/lib/uploads
```

On `compose` this creates a named Docker volume; on `k3s` it creates a PVC
backed by Longhorn (or your default StorageClass).

## Validation

Blueprints are validated on create and at deploy time:

- `name` is required and services must be unique
- at least one service is required
- `build.type` must be `dockerfile`, `buildpacks` or `static`
- `autoscaling.minReplicas` cannot exceed `maxReplicas`
- persistence entries require `name` and `mountPath`

## Examples

- [examples/custodian.yaml](../examples/custodian.yaml) — full-featured example
- [Quickstart in the README](../README.md#quickstart) — minimal "hello world"
