#!/usr/bin/env bash
set -euo pipefail
CLUSTER="${KIND_CLUSTER_NAME:-spotsync}"
kind delete cluster --name "$CLUSTER" 2>/dev/null || true
