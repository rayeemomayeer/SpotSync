# Observability

## Signals

| Signal | Where | Notes |
| --- | --- | --- |
| Structured logs | Render / stdout (`slog`) | Correlation via `X-Request-Id` |
| Prometheus metrics | `GET /metrics` | Gate with `METRICS_TOKEN` in prod |
| OpenTelemetry traces | OTLP HTTP | Optional: `OTEL_EXPORTER_OTLP_ENDPOINT` |
| SLOs | [grafana/slo.md](./grafana/slo.md) | Availability, reserve p95, oversell=0 |
| Web client errors | spotsync-web | Optional `NEXT_PUBLIC_SENTRY_DSN` stub |

## Enable tracing locally

```bash
# Example: Jaeger all-in-one
docker run -d --name jaeger -p 4318:4318 -p 16686:16686 jaegertracing/all-in-one:1.57

export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
# then start the API — logs "otel tracing enabled" when wired
```

## Dashboards

See [grafana/README.md](./grafana/README.md) and local Prometheus via compose (`:9090`).
