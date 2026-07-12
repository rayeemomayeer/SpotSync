# Incident response (portfolio)

## Severity

| Level | Example | Action |
| --- | --- | --- |
| SEV1 | Over-sell detected / data loss | Freeze deploys, set zones read-only if needed, restore from Neon |
| SEV2 | API down / ready failing | Check Render logs, Neon, Redis; rollback last deploy |
| SEV3 | SSE stale / notify lag | Check worker, Redis, outbox dead-letter rows |

## First 15 minutes

1. Capture `X-Request-Id` from failing client.
2. Hit `/healthz` and `/readyz`.
3. Check Render deploy timeline + Neon status.
4. Scan `/metrics` (with `METRICS_TOKEN`) for latency / oversell counters.
5. If auth abuse: confirm Redis rate limit keys and 429 responses.

## Comms template

```
Impact: ...
Start: ...
Status: investigating | mitigated | resolved
Next update: ...
```
