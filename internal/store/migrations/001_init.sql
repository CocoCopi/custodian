-- Custodian control plane schema (v1).
-- Applied idempotently at startup; safe to re-run.

CREATE TABLE IF NOT EXISTS services (
    id          TEXT PRIMARY KEY,
    owner_id    TEXT NOT NULL,
    name        TEXT NOT NULL,
    repo_url    TEXT NOT NULL DEFAULT '',
    branch      TEXT NOT NULL DEFAULT 'main',
    build_type  TEXT NOT NULL DEFAULT 'dockerfile',
    image       TEXT NOT NULL DEFAULT '',
    blueprint   TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'provisioning',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (owner_id, name)
);

CREATE TABLE IF NOT EXISTS deployments (
    id          TEXT PRIMARY KEY,
    service_id  TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    commit_sha  TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'provisioning',
    image       TEXT NOT NULL DEFAULT '',
    logs        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_deployments_service ON deployments (service_id, created_at DESC);

CREATE TABLE IF NOT EXISTS api_tokens (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    owner_id    TEXT NOT NULL,
    token_hash  TEXT NOT NULL,
    prefix      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);
