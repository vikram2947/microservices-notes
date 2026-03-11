## Module 3 – Communication & Reliability Demo

This module runs two services in a single Go binary:

- `slow-service` (downstream) on port `SLOW_PORT` (default **8091**).
- `api service` (client) on port `API_PORT` (default **8090**).

### Run

```bash
cd module3-comm

export API_PORT=8090
export SLOW_PORT=8091
export QUEUE_WORKERS=2

go run ./...
```

### 1) Synchronous call with timeout, retry, circuit breaker, and rate limiting

Endpoint:

```bash
curl -i http://localhost:8090/do-work
```

What happens:

- API service calls `slow-service /work` with:
  - HTTP client timeout.
  - Up to 3 attempts with exponential backoff + jitter.
  - Circuit breaker that opens after a few failures and returns `503` quickly.
- Rate limiter allows a limited number of requests per time window; beyond that you get `429`.

Try:

- Send multiple `curl` calls quickly and watch logs and responses (`200`, `502/503`, `429`).

### 2) Asynchronous job queue with retries (at-least-once style)

Enqueue a job:

```bash
curl -s -X POST http://localhost:8090/enqueue \
  -H 'Content-Type: application/json' \
  -d '{"data":"do something important"}'
```

You’ll get:

```json
{"job_id":"job_xxx"}
```

Check job status:

```bash
curl -s http://localhost:8090/jobs/job_xxx
```

Possible statuses: `"DONE"` or `"FAILED"` (after a few retries).

This simulates:

- Producer (`/enqueue`) not blocked on actual work.
- Background workers processing with retries and backoff.

