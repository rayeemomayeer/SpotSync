# Prometheus scrape config

Local Prometheus targets the compose `api` service. Production scrapes Render `/metrics` (protect with `METRICS_TOKEN`).

See [grafana/README.md](../grafana/README.md) and [observability.md](../observability.md).
