#!/usr/bin/env bash
set -euo pipefail
CLUSTER="${KIND_CLUSTER_NAME:-spotsync}"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
K8S="$ROOT/deploy/k8s"

if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER"; then
  kind create cluster --name "$CLUSTER"
fi

docker compose -f "$ROOT/deploy/compose/docker-compose.yml" build api worker
kind load docker-image spotsync-api:latest --name "$CLUSTER"
kind load docker-image spotsync-worker:latest --name "$CLUSTER"

kubectl apply -f "$K8S/namespace.yaml"
kubectl apply -f "$K8S/postgres-redis.yaml"
if [[ ! -f "$K8S/secrets.yaml" ]]; then
  cp "$K8S/secrets.example.yaml" "$K8S/secrets.yaml"
fi
kubectl apply -f "$K8S/secrets.yaml"
kubectl apply -f "$K8S/spotsync.yaml"

kubectl -n spotsync rollout status deployment/spotsync-postgres --timeout=120s
kubectl -n spotsync rollout status deployment/spotsync-redis --timeout=120s
kubectl -n spotsync rollout status deployment/spotsync-api --timeout=180s
kubectl -n spotsync rollout status deployment/spotsync-worker --timeout=180s

echo "SpotSync kind cluster ready."
echo "Port-forward: kubectl -n spotsync port-forward svc/spotsync-api 8080:80"
