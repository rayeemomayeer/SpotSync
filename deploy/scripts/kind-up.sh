#!/usr/bin/env bash
set -euo pipefail
CLUSTER="${KIND_CLUSTER_NAME:-spotsync}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  kind create cluster --name "$CLUSTER"
fi

docker compose -f "$ROOT/deploy/compose/docker-compose.yml" build api worker
kind load docker-image spotsync-api:latest --name "$CLUSTER" 2>/dev/null || true
kind load docker-image spotsync-worker:latest --name "$CLUSTER" 2>/dev/null || true

kubectl apply -f "$ROOT/deploy/k8s/spotsync.yaml"
echo "SpotSync kind cluster ready. Port-forward: kubectl -n spotsync port-forward svc/spotsync-api 8080:80"
