# SpotSync runbook

Operational guide for the API, worker, and production stack.

## Live stack

| Component | URL / service |
| --- | --- |
| **API** | https://spotsync-ei6g.onrender.com |
| **Frontend** | https://spotsync-nu.vercel.app |
| **API repo** | https://github.com/rayeemomayeer/SpotSync |
| **Frontend repo** | https://github.com/rayeemomayeer/spotsync-web |
| **Database** | Neon PostgreSQL (pooled `DATABASE_URL`) |
| **Cache / SSE fan-out** | Redis (`REDIS_URL`, optional — Upstash recommended on Render) |

## Health checks

```bash
curl https://spotsync-ei6g.onrender.com/healthz   # process up
curl https://spotsync-ei6g.onrender.com/readyz    # Postgres reachable
curl https://spotsync-ei6g.onrender.com/metrics   # Prometheus scrape
```

## Deploy (Render)

1. Push to `main` — Render auto-deploys the web service from GitHub.
2. **Worker:** apply `render.yaml` Blueprint or add `spotsync-worker` manually (same env as API minus CORS).
3. Set secrets in Render dashboard:
   - `DATABASE_URL`, `DATABASE_MIGRATE_URL` (Neon non-pooled host for migrations)
   - `JWT_SECRET`
   - `REDIS_URL` (optional; enables cross-replica SSE + availability cache)
   - `CORS_ALLOWED_ORIGINS` — `https://spotsync-nu.vercel.app` (comma-separate for previews)
4. Manual GitHub deploy: Actions → **Deploy Render** (needs `RENDER_API_KEY`, `RENDER_SERVICE_ID`).

### Post-deploy seed (first time or empty DB)

```bash
DATABASE_URL="your-neon-url" go run ./cmd/seed
```

## Roll back

1. Render dashboard → service → **Rollback** to previous deploy.
2. If a bad migration shipped, run `migrate down 1` against Neon with `DATABASE_MIGRATE_URL`, then redeploy fixed code.

## Scale

- **Render free tier:** single API instance; in-process SSE hub works without Redis.
- **Multiple API replicas:** set `REDIS_URL` so `BridgeRedis` fans out SSE and the worker relays outbox events.
- **Worker:** one replica is enough; runs outbox relay + scheduled expiry loops.

## Observe

| Signal | Where |
| --- | --- |
| Request logs | Render logs (structured `slog`, `X-Request-Id`) |
| `reservation_latency_seconds` | `/metrics` (requires `Authorization: Bearer $METRICS_TOKEN` or `?token=` when set) |
| `oversell_attempts_rejected_total` | `/metrics` |
| `zone_availability_cache_hits_total` | `/metrics` (when Redis configured) |
| Local stack | `docker compose -f deploy/compose/docker-compose.yml up` + Prometheus on `:9090` |

SSE zone streams require a JWT (`Authorization: Bearer` or `?access_token=`). Set `ALLOW_SELF_ADMIN_REGISTRATION=false` in production and seed admins. Set `METRICS_TOKEN` on Render. Worker + `REDIS_URL` (Upstash free) recommended for outbox relay and multi-replica SSE.

## Load test (k6)

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
go run ./cmd/seed
API_BASE=http://localhost:8080/api/v1 k6 run test/load/reserve_stampede.js
```

## Capacity strategies

`CAPACITY_STRATEGY` accepts `row_lock` (default), `optimistic`, `redis_counter`. All three are implemented; `redis_counter` requires `REDIS_URL` and uses Postgres row lock as final arbiter after Redis fast-reject.

## Kubernetes (local / CI)

See `deploy/k8s/README.md` — kind manifests for multi-replica API + worker; not used in production Render deploy.
