# Security Policy

Custodian takes security seriously. The control plane holds the keys to your
deployments — please report vulnerabilities privately and responsibly.

## Reporting a vulnerability

**Do not open a public issue for security problems.** Email the maintainers at
security@example.com (or use GitHub's private vulnerability reporting: *Security
→ Report a vulnerability*).

Please include:

- Affected version(s) and component
- A minimal reproduction
- Impact assessment, if known

We aim to acknowledge reports within 48 hours and to ship a fix (or a
mitigation plan) within 14 days.

## Supported versions

| Version | Supported |
| --- | --- |
| latest `main` | ✅ |
| tagged releases | ✅ for 90 days after the next release |
| older releases | ❌ |

## Security posture

- Secrets are injected at deploy time and never persisted in the database.
- API token plaintext is shown once; only SHA-256 hashes are stored.
- All API routes except `/healthz` and OIDC endpoints require authentication.
- TLS is terminated by Traefik with automatic Let's Encrypt certificates.
- Keep `CUSTODIAN_JWT_SECRET` strong (`openssl rand -hex 32`) and rotate
  regularly.

## Known hardening steps for production

1. Restrict the control plane host's Docker socket access.
2. Put Cloudflare (or another CDN/WAF) in front of Traefik for DDoS
   protection and caching.
3. Enable fail2ban or equivalent on the host for SSH.
4. Run regular Postgres backups and test restores.
