## Module 7 – Observability: Logging, Metrics, Tracing

### 1. Observability pillars (logs, metrics, traces)

- **Logs**:
  - Structured, machine-parsable logs (JSON or key-value).
  - Include context: request IDs, user IDs, service name.
- **Metrics**:
  - Numeric time series: counters, gauges, histograms.
  - Used for dashboards, alerts, SLOs.
- **Traces**:
  - End-to-end view of a request across services.
  - Spans with timing and attributes; propagate a trace ID.

Goal: be able to answer “What is happening?” and “Why is it slow/failing?” across microservices.

---

### 2. Correlation IDs and context propagation

- **Correlation ID / Trace ID**:
  - A unique identifier added to a request when it enters the system.
  - Propagated via headers (e.g., `X-Request-ID`, `traceparent`).
  - Logged and attached to metrics.
- In Go:
  - Store ID in `context.Context`.
  - Middleware:
    - Extract or generate ID.
    - Put into context.
    - Add to outbound calls as a header.

---

### 3. Metrics: RED and USE

- **RED (for services)**:
  - Rate: requests per second.
  - Errors: error rate.
  - Duration: latency (usually percentiles).
- **USE (for resources)**:
  - Utilization, Saturation, Errors for things like CPU, memory, queues.

In this module:

- We’ll implement a simple `/metrics` endpoint with:
  - Request counts per route.
  - Simple latency histograms.

---

### 4. Tracing (OpenTelemetry-lite intro)

- **Trace**: a tree of spans representing a request journey.
- **Span**: a single operation (HTTP handler, DB query, external call).
- **Propagation**:
  - Use standard headers (`traceparent`) so different libraries/services interoperate.

For simplicity here:

- We’ll implement:
  - A minimal **trace ID** propagation using `X-Trace-ID`.
  - Log that ID in every service.
  - Show how you would hook in OpenTelemetry in real systems.

---

### 5. What we implement in Module 7

Single Go project `module7-observability` that:

- Wraps a small HTTP service with:
  - Logging middleware that:
    - Extracts/creates `X-Trace-ID`.
    - Logs structured lines including `trace_id`, method, path, status, latency.
  - Metrics:
    - In-memory counters for total requests and errors per path.
    - Basic latency buckets per path.
    - `GET /metrics` endpoint exposing a simple text format.
- Uses `X-Trace-ID` when calling an internal “downstream” handler:
  - Simulates how trace IDs propagate across services.

