# Docker

Multi-stage image for the SpotSync API (`cmd/api`).

```bash
docker build -f deploy/docker/Dockerfile -t spotsync-api .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="dev-secret" \
  spotsync-api
```

Production deploys via Fly.io (`fly.toml` at repo root). See [README](../../README.md#deployment).
