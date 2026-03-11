## Module 11 – Performance & Caching Demo

This module compares:

- `/data/raw` – no cache, always hits slow source.
- `/data/cached` – in-memory cache with TTL over the same slow source.

It also includes a simple built-in **load generator** at `/bench`.

### Run

```bash
cd module11-perf

# Optional: override port (default 8700)
export APP_PORT=8700

go run ./...
```

### 1. Try endpoints manually

Raw (no cache):

```bash
curl -s "http://localhost:8700/data/raw?key=test"
```

Cached:

```bash
curl -s "http://localhost:8700/data/cached?key=test"
```

Call the cached endpoint multiple times with the same key; you should see:

- First call: `"source": "slow"`.
- Subsequent calls within TTL: `"source": "cache"` and much smaller delay.

### 2. Run benchmark

Without cache:

```bash
curl -s "http://localhost:8700/bench?target=raw&n=200"
```

With cache:

```bash
curl -s "http://localhost:8700/bench?target=cached&n=200"
``>

You’ll get JSON with:

- `avg_ms`, `p95_ms` – basic latency summary.
- `cache_hits`, `cache_misses` – cache effectiveness.

Compare raw vs cached to see how caching improves latency and reduces load on the slow source.

