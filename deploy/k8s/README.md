# Kubernetes (Phase 5)

Manifests for local **kind** clusters and CI e2e — production uses Render (see root `render.yaml`).

## Contents

- `spotsync.yaml` — API Deployment (2 replicas), worker Deployment, Services, HPA stub

## Quick start (kind)

```bash
kind create cluster --name spotsync
docker build -f deploy/docker/Dockerfile -t ghcr.io/rayeemomayeer/spotsync-api:latest .
docker build -f deploy/docker/Dockerfile.worker -t ghcr.io/rayeemomayeer/spotsync-worker:latest .
kind load docker-image ghcr.io/rayeemomayeer/spotsync-api:latest --name spotsync
kind load docker-image ghcr.io/rayeemomayeer/spotsync-worker:latest --name spotsync
kubectl apply -f deploy/k8s/spotsync.yaml
kubectl -n spotsync create secret generic spotsync-secrets \
  --from-literal=DATABASE_URL=... \
  --from-literal=JWT_SECRET=...
```

## Notes

- Set `REDIS_URL` in `spotsync-secrets` when running more than one API replica.
- Rolling updates: `kubectl rollout status deployment/spotsync-api -n spotsync`
- CI kind e2e is optional; integration tests in GitHub Actions cover the capacity invariant today.
