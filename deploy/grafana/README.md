# Grafana dashboards (Phase 4)

SpotSync exposes Prometheus metrics at `GET /metrics`. Import or build dashboards from these series:

| Metric | Type | Use |
| --- | --- | --- |
| `reservation_latency_seconds` | histogram | p50/p95 reserve latency (`method=POST`) |
| `oversell_attempts_rejected_total` | counter | 409 conflict rate under load |
| `zone_availability_cache_hits_total` | counter | Redis cache effectiveness |
| `zone_availability_cache_misses_total` | counter | Cache warm-up / invalidation |

## Local

1. `docker compose -f deploy/compose/docker-compose.yml up prometheus api`
2. Open http://localhost:9090
3. Query `rate(oversell_attempts_rejected_total[5m])`

## Production

Scrape `https://spotsync-ei6g.onrender.com/metrics` from Grafana Cloud or self-hosted Prometheus (protect with network rules).

Suggested panels:

- **Reserve latency** — `histogram_quantile(0.95, sum(rate(reservation_latency_seconds_bucket[5m])) by (le))`
- **Oversell rejections** — `increase(oversell_attempts_rejected_total[1h])`
- **Cache hit ratio** — `rate(zone_availability_cache_hits_total[5m]) / (rate(zone_availability_cache_hits_total[5m]) + rate(zone_availability_cache_misses_total[5m]))`

OpenTelemetry tracing is deferred; logs + Prometheus cover the current production stack.
