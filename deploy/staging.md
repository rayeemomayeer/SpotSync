# Free-tier staging

## Topology

| Piece | Staging approach |
| --- | --- |
| Postgres | Neon branch `staging` |
| Redis | Upstash free DB (separate from prod if quota allows) |
| Go API + worker | Render free services (`spotsync-staging`, `spotsync-staging-worker`) |
| BFF | Render free `spotsync-bff-staging` |
| Notify | Render free `spotsync-notify-staging` |
| Web | Vercel Preview / staging project |

## Promote path

1. Migrate staging Neon branch (`MIGRATE_ON_STARTUP=true`).
2. Seed staging (`SEED_ADMIN_*`).
3. Deploy API → worker → BFF → notify → web preview.
4. Smoke: `/healthz`, `/readyz`, login, reserve, SSE with JWT.

## Backup / restore drill

1. Neon console → project → Branches → create branch from PITR.
2. Point a throwaway Render service at the restore URL.
3. Confirm `/readyz` and `GET /api/v1/zones`.
4. Record RPO/RTO notes in the runbook after each drill.

## Kind

`deploy/k8s` + kind scripts remain **local/CI portfolio proof only** — not a paid cluster.
