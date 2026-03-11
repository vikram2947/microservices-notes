## Module 3 – Communication Patterns & Reliability (Sync + Async)

### 1. Synchronous communication (HTTP) – patterns

- **Request–response**:
  - Caller sends a request and waits for a response (`GET /orders/{id}`).
  - Simple mental model, but couples latency and availability of both sides.
- **Aggregator / API composition**:
  - One service calls multiple downstream services and aggregates results for the client.
- **Fan-out / fan-in**:
  - One request triggers parallel calls to multiple services; results are combined.

Key rule: **Each network call must have a timeout and usually a retry policy.**

---

### 2. Asynchronous communication – patterns

- **Message / event-driven**:
  - Publisher sends a message to a queue or topic.
  - Consumers process messages independently of the publisher.
- **Pub/sub**:
  - One message can be consumed by multiple subscribers.
- **Competing consumers**:
  - Multiple instances consume from the same queue to scale out.

Benefits:

- Decouples caller from callee availability.
- Can absorb spikes via buffering.

Tradeoff: harder to reason about immediate state; usually **eventual consistency**.

---

### 3. Reliability patterns (critical for interviews)

- **Timeouts**:
  - Never rely on default infinite/blocking operations.
  - Set explicit timeouts on HTTP clients and any other I/O.
- **Retries with backoff + jitter**:
  - Retries help with transient failures (network blips, brief outages).
  - Backoff (e.g., exponential) prevents hammering a struggling service.
  - Jitter randomizes timing to avoid synchronized retry storms.
- **Circuit breaker**:
  - Monitors failures/latency of a downstream.
  - When failures exceed a threshold, the circuit “opens” and short-circuits future calls (fast error) for a while.
  - After a cool-off, tries a limited number of requests (“half-open”).
- **Bulkhead**:
  - Isolates resources so one failing dependency doesn’t exhaust all threads/connections.
  - In Go, often implemented with bounded worker pools and separate clients.
- **Rate limiting**:
  - Protects your service or downstreams by limiting requests per unit time.
  - Simple implementations: token bucket, leaky bucket.
- **Load shedding**:
  - When under heavy load, reject or shed non-critical requests early instead of failing everything.

---

### 4. Idempotency & deduplication

- With retries and at-least-once delivery, a server may see the **same logical request more than once**.
- **Idempotent operation**:
  - Applying the same operation multiple times has the same effect as once.
  - Example: `PUT /users/{id}` to set data vs `POST /charges` that creates a new charge each time.
- **Idempotency keys**:
  - Client sends a unique key (e.g., `Idempotency-Key` header or request ID).
  - Server stores result per key and returns the same result when the same key is reused.
- **Consumer-side dedup**:
  - For messages, track processed message IDs and skip repeats.

---

### 5. What we implement in Module 3

One small Go project with:

- **Downstream service** (`slow-service`):
  - Endpoint `/work` that sometimes succeeds, sometimes fails or is slow.
- **Client/API service** (`api-gateway-like`):
  - HTTP endpoint `/do-work` that calls `slow-service`.
  - Uses:
    - HTTP client with timeout.
    - Retry with exponential backoff + jitter.
    - Simple in-memory circuit breaker.
    - Simple rate limiter middleware.
- **Async worker simulation**:
  - In-memory queue (`chan`) where `/enqueue` endpoint accepts a job.
  - Background worker processes jobs with retries and idempotent handling.

You should observe:

- How timeouts protect the client from hanging.
- How retries + backoff handle transient failures.
- How circuit breaker avoids hammering a broken downstream.
- How an async queue decouples producers from workers.

