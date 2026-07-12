# Local compose stack

Postgres, Redis, API, worker, nginx, Prometheus for local multi-replica experiments.

```bash
docker compose -f deploy/compose/docker-compose.yml up -d --build
```

## Sibling Node services

Run from their repos (not this compose file):

| Service | Repo | Default port | Notes |
| --- | --- | --- | --- |
| BFF (Better Auth + proxy) | `spotsync-bff` | 4000 | `GO_API_BASE_URL` → Go; optional `NOTIFY_URL` |
| Notify (Resend + Redis) | `spotsync-notify` | 3100 | Same `REDIS_URL` as API; `INTERNAL_TOKEN` shared with BFF |
| Web | `spotsync-web` | 3000 | Vercel locally via `npm run dev` |

Suggested local wiring:

```text
web :3000  →  BFF :4000  →  Go API :8080 (nginx) / :8081 (direct)
                 ↓
           notify :3100  ←  Redis pub/sub from API/worker
```

- BFF `GO_API_BASE_URL`: `http://localhost:8080` (or `http://localhost:8081`)
- BFF `FRONTEND_ORIGIN`: `http://localhost:3000`
- BFF `NOTIFY_URL`: `http://localhost:3100` (optional)
- Notify `REDIS_URL`: `redis://localhost:6379` (same as compose Redis)
- Web `NEXT_PUBLIC_BFF_URL`: `http://localhost:4000`
- Web `NEXT_PUBLIC_API_BASE_URL`: `http://localhost:8081/api/v1` (or via BFF proxy)

Staging topology: [../staging.md](../staging.md). Observability: [../observability.md](../observability.md).
