## Module 7 – Observability Demo (Logs, Metrics, Trace IDs)

This module shows:

- Structured logging with `trace_id`.
- Basic RED-style metrics per route.
- Simple trace ID propagation via `X-Trace-ID`.

### Run

```bash
cd module7-observability

# Optional: override port (default 8500)
export APP_PORT=8500

go run ./...
```

Service listens on `http://localhost:8500`.

### 1. Try the `/work` endpoint

```bash
curl -s http://localhost:8500/work
```

or with a custom trace ID:

```bash
curl -s http://localhost:8500/work -H "X-Trace-ID: my-trace-123"
```

Observe:

- Response JSON includes `trace_id` and `latency`.
- Logs in the terminal include `trace_id=...` for:
  - Incoming request.
  - Downstream simulated call.
  - Completion line.

### 2. Check metrics

```bash
curl -s http://localhost:8500/metrics
```

You’ll see per-route metrics:

- `requests_total`
- `errors_total`
- `latency_bucket{le="..."}` counts

This is similar in spirit to Prometheus metrics and supports RED-style monitoring.

