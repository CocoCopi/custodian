# Contributing to Custodian

Thanks for helping make self-hosted PaaS better. Every contribution — code,
docs, bug reports, design feedback — is valued.

## Code of Conduct

By participating you agree to uphold our [Code of Conduct](CODE_OF_CONDUCT.md).
Be excellent to each other.

## Getting started

1. **Fork** the repository and clone your fork.
2. Set up the dev environment:
   ```bash
   make build        # builds custodian-api, custodian-worker and custodian
   make test         # runs the backend test suite
   make frontend-build
   ```
3. Create a branch: `git checkout -b feat/your-feature`.

## What to work on

- Check the [Roadmap](README.md#roadmap) and open issues for feature work.
- Good first issues are labelled `good first issue`.

## Development workflow

- Backend code lives in `internal/`, entrypoints in `cmd/`.
- Add tests for any new logic (blueprint validation, deployer rendering,
  auth, store).
- Run `make vet` and `make lint` before pushing.
- Frontend changes go in `frontend/`; run `npm run typecheck` and
  `npm run build`.
- Keep commits focused and use clear messages:
  `feat(blueprint): support scaleToZero`, `fix(deployer): correct ingress path`.

## Pull requests

1. Ensure CI passes (lint, vet, tests, frontend build).
2. Update `docs/` when behaviour changes.
3. Reference the issue you're closing: `Closes #123`.

## Project layout

```
cmd/              entrypoints (custodian-api, custodian-worker, custodian-cli)
internal/api      REST + WebSocket handlers
internal/auth     OIDC + API tokens + middleware
internal/blueprint  custodian.yaml parsing/validation
internal/deployer   compose + k8s artifact rendering
internal/jobs     Asynq build/apply workers
internal/store    PostgreSQL persistence (embedded migrations)
internal/ws       WebSocket log hub
frontend/         React + TypeScript console
deploy/           docker-compose, k3s and observability manifests
docs/             architecture, blueprint, API, deployment guides
```

## License

By contributing you agree that your contributions are licensed under the
[Apache License 2.0](LICENSE).
