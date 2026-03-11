## Module 11 – Performance, Scalability & Caching

### 1. Core performance concepts

- **Throughput**: requests per second your service handles.
- **Latency**: time to handle a single request (look at percentiles: p50, p95, p99).
- **Tail latency**: high-percentile latencies (p95/p99) that users feel as “slowness”.
- **Capacity**: max load before error rates/latency spike.

Rule: always think in terms of **SLOs** (e.g., “p95 latency < 200ms for 99% of requests”).

---

### 2. Load patterns & scaling

- **Steady** vs **spiky** traffic.
- Scaling options:
  - Vertical (bigger machine).
  - Horizontal (more instances).
- In microservices:
  - Horizontal scaling is common: more Pods/containers behind a Service.
  - Use autoscaling based on CPU/RPS/queue depth.

---

### 3. Caching strategies

- **Where to cache**:
  - In-process (in-memory) cache: fastest, per-instance.
  - Shared cache: Redis/Memcached; shared across instances.
  - Client-side / CDN for static or cacheable responses.
- **Key questions**:
  - What can be stale (and for how long)?
  - What invalidation strategy do you need?
- Simple patterns:
  - Cache-aside (lazy loading).
  - Time-based eviction (TTL).

---

### 4. Queues & buffering

- Use queues to:
  - Smooth spikes (buffer requests and drain with workers).
  - Offload slow or non-critical work from sync path.
- Metrics to watch:
  - Queue length.
  - Processing rate.
  - Time jobs spend in queue.

---

### 5. Go performance basics

- Avoid unnecessary allocations.
- Reuse HTTP clients (connection pooling).
- Use `pprof` for CPU/heap profiling:
  - Add `/debug/pprof` and run `go tool pprof`.

---

### 6. What we implement in Module 11

Single Go project `module11-perf` that:

- Implements:
  - A **read-heavy endpoint** with in-memory cache (TTL) over a simulated slow data source.
  - An endpoint **without** cache for comparison.
- Includes:
  - Simple built-in **load generator** to hit endpoints.
  - Logging of average and p95 latency for cached vs non-cached paths.

