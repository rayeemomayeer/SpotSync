# github.com/rayeemomayeer/SpotSync/deploy/grafana/slo.md — free-tier SLO notes

## Service level objectives (portfolio)

| SLI | Target | Measurement |
| --- | --- | --- |
| API availability | 99% monthly (best-effort free tier) | `/readyz` success ratio |
| Reserve p95 latency | < 500ms local / < 2s free Render | `reservation_latency_seconds` |
| Oversell | 0 successful oversells | stampede tests + `oversell_attempts_rejected_total` |
| Outbox lag | < 30s under load | worker processed_at vs created_at |

## Alerts (manual / Grafana)

- `/readyz` failing 3m
- Outbox dead-letter count increasing
- Auth 429 spike (abuse)

## Tracing

Set `OTEL_EXPORTER_OTLP_ENDPOINT` (e.g. local Jaeger `http://localhost:4318`) to enable OTLP export from the Go API.
