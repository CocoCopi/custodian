# REST API

Base URL: `https://api.<your-domain>` (dev: `http://localhost:8080`)

Authentication: `Authorization: Bearer <session-jwt | cst_* api token>`

## Health

```
GET /healthz
```

```json
{ "status": "ok", "engine": "compose", "version": "0.1.0", "oidc": true }
```

## Identity

```
GET /api/v1/me
```

```json
{ "owner_id": "user-123" }
```

## Services

### List

```
GET /api/v1/services
```

```json
{
  "services": [
    {
      "id": "…",
      "owner_id": "user-123",
      "name": "hello",
      "repo_url": "",
      "branch": "main",
      "build_type": "dockerfile",
      "status": "running",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

### Create

```
POST /api/v1/services
```

```json
{
  "name": "hello",
  "repo_url": "https://github.com/you/hello.git",
  "build_type": "dockerfile",
  "blueprint": "apiVersion: custodian.dev/v1\nkind: Blueprint\nname: hello\nservices: […]"
}
```

Returns `201` with the created service. `409` if the name is taken.

### Get / Delete

```
GET    /api/v1/services/:id
DELETE /api/v1/services/:id   → 204
```

## Deployments

### Trigger a deploy

```
POST /api/v1/services/:id/deployments?commit=<sha>
```

Returns `202` with the queued deployment. The build and apply jobs run
asynchronously; status updates are visible via `GET` below and logs via the
WebSocket endpoint.

### List / Get

```
GET /api/v1/services/:id/deployments
GET /api/v1/deployments/:id
```

## Live logs (WebSocket)

```
WS  /api/v1/services/:id/logs       # stream by service
WS  /api/v1/deployments/:id/logs    # stream by deployment
```

Each frame is a JSON log entry:

```json
{
  "deployment_id": "…",
  "service_id": "…",
  "stream": "stdout",
  "message": "building image for web@main\n",
  "timestamp": "2026-01-01T00:00:00Z"
}
```

## API tokens

```
GET    /api/v1/tokens           # list
POST   /api/v1/tokens           # {"name": "ci"} → plaintext returned once
DELETE /api/v1/tokens/:id       # revoke
```

## OIDC

```
GET /api/v1/auth/login          # redirect to the IdP
GET /api/v1/auth/callback       # exchange code, issue session JWT
```
