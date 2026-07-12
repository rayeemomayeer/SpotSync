# k6 load scripts

Stampede reserve scenario used in CI against the compose stack.

```bash
API_BASE=http://localhost:8080/api/v1 k6 run test/load/reserve_stampede.js
```

Additional read-heavy scenarios can share the same `API_BASE` pattern against staging.
