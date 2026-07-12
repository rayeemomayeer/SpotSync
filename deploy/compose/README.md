# Local compose stack

Postgres, Redis, API, worker, nginx, Prometheus for local multi-replica experiments.

```bash
docker compose -f deploy/compose/docker-compose.yml up -d --build
```

Sibling Node services (run from their repos, not this compose file):

| Service | Repo | Default port |
| --- | --- | --- |
| BFF (Better Auth + proxy) | `spotsync-bff` | 4000 |
| Notify (Resend + Redis) | `spotsync-notify` | 4100 |

Point BFF `GO_API_BASE_URL` at `http://localhost:8080/api/v1` (or nginx `:8080`) and `FRONTEND_ORIGIN` at `http://localhost:3000`.
