## Module 13 – Pattern Catalog & Interview Drilldown

### 1. Pattern catalog (you should be able to explain each)

For each, you should know:

- **What it is**
- **When to use it**
- **Key tradeoffs**
- **(Ideally) where we implemented or simulated it in this repo**

#### Decomposition & design

- **Bounded Context**
  - One cohesive domain model with its own language and rules.
  - Use to define microservice boundaries.
  - Seen in: `users`, `orders`, `payments` in Module 2.

- **Strangler Fig**
  - Gradually replace parts of a legacy system by routing specific features to new services.
  - Use when migrating a monolith to microservices without big-bang rewrite.

- **Anti-Corruption Layer**
  - A translation layer between a clean domain model and an external/legacy system.
  - Use to avoid polluting your model with legacy semantics.

#### Communication

- **API Gateway**
  - Single entry point for clients; routes to services, handles edge concerns.
  - Implemented in: `module5-gateway`.

- **Backend for Frontend (BFF)**
  - Client-specific backend that aggregates data from multiple services.
  - Implemented conceptually via `/me/overview` in `module5-gateway`.

- **Aggregator**
  - Service that calls multiple downstream services and returns combined result.
  - BFF is a common form of aggregator.

- **Publish/Subscribe**
  - Producers publish events; multiple consumers can subscribe.
  - Simulated via in-memory event bus in `module4-saga`.

- **Competing Consumers**
  - Multiple workers read from the same queue to scale processing.
  - Simulated via multiple workers in `module3-comm` job queue.

#### Data & consistency

- **Database-per-service**
  - Each service owns its own schema/data store.
  - Simulated via separate in-memory stores in Module 2.

- **Saga (orchestration & choreography)**
  - Sequence of local transactions with compensations across services.
  - Orchestration-style demo: `module4-saga` (payment + order status).

- **Outbox pattern**
  - Write domain data + outbox event in same local transaction; later publish events from outbox.
  - Simulated via `OutboxStore` in `module4-saga`.

- **CQRS**
  - Separate write model (commands) from read model (queries).
  - Demo: `OrderStore` vs `SummaryStore` in `module4-saga`.

- **Event Sourcing (concept)**
  - Persist events as the source of truth; rebuild state from events.
  - You can point to `Event` model in `module4-saga` as a first step.

- **Materialized View**
  - Precomputed read model built from events (like `order_summary`).
  - Implemented via `SummaryStore` in `module4-saga`.

- **Idempotent Consumer**
  - Consumer that can process the same message multiple times safely.
  - Our job worker in `module3-comm` is designed to handle retries without double-side-effects.

#### Resilience

- **Timeout**
  - Always set upper bounds on remote calls.
  - Implemented in HTTP clients in `module3-comm` and other modules.

- **Retry with Exponential Backoff & Jitter**
  - Retry transient failures, increasing delay with randomness to avoid thundering herds.
  - Implemented in `callSlowServiceWithRetry` and worker backoff in `module3-comm`.

- **Circuit Breaker**
  - Trip open after failures; fail fast; half-open to probe recovery.
  - Implemented via `circuitBreaker` in `module3-comm`.

- **Bulkhead**
  - Isolate resources so one slow dependency doesn’t exhaust all capacity.
  - Concept shown via separate worker pools/queues; in code, multiple workers limited via buffered channels.

- **Rate Limiter**
  - Limit requests per unit time to protect services.
  - Implemented via token bucket in `module3-comm` and `module5-gateway`.

- **Load Shedding**
  - Reject or drop low-priority work under high load.
  - Simulated via queue full path in `/enqueue` (`module3-comm`).

- **Dead Letter Queue**
  - Store messages that repeatedly fail processing for later inspection.
  - Conceptual; we mark jobs as `FAILED` after retries in `module3-comm`.

#### Operations & delivery

- **Service Discovery**
  - Find service instances dynamically via registry or DNS.
  - Simulated registry in `module6-discovery`.

- **Config Externalization**
  - Use env/config files instead of hardcoding values.
  - Used across modules (e.g., `APP_PORT`, config.json in Module 6).

- **Health Checks**
  - `/healthz`, `/readyz` endpoints used throughout (Modules 1, 2, 3, etc.).

- **Blue/Green & Canary**
  - Two versions side-by-side; shift traffic gradually (canary).
  - Demonstrated conceptually with Istio `VirtualService` in Module 12.

- **Feature Flags**
  - Toggle features on/off without redeploy; often config-driven or via flag service.

#### Observability

- **Correlation IDs / Trace IDs**
  - IDs propagated across services for logs/traces.
  - Implemented via `X-Trace-ID` and context in `module7-observability`.

- **Structured Logging**
  - Key-value or JSON logs for machine parsing.
  - Used throughout with log prefixes and fields.

- **Distributed Tracing**
  - Traces/spans across calls; conceptually mirrored with `trace_id` logs in `module7-observability`.

- **Metrics & SLIs**
  - RED metrics (Rate, Errors, Duration) in `module7-observability`.

#### Security

- **mTLS**
  - Mutual TLS between services; implemented at mesh layer in real systems.
  - Conceptual Istio config in Module 12.

- **Token Relay**
  - Gateway validates token, forwards identity to backend via headers.
  - Implemented in `module8-security`.

- **Zero-Trust-ish**
  - Assume network is hostile; authenticate/authorize every call.
  - Combined effect of tokens, mTLS, and mesh policies.

#### Testing

- **Contract Testing**
  - Ensure API shape/behavior meets consumer expectations.
  - Example: `/healthz` contract test in `module1-service/main_test.go`.

- **Chaos Testing (concept)**
  - Intentionally break things (kill pods, inject latency) to validate resilience.

---

### 2. Interview drill – core question templates

You should practice answering these out loud in 3–5 minutes each.

- **“Design an order management system with microservices.”**
  - Talk about services: `orders`, `payments`, `inventory`, `users`.
  - Boundaries and data ownership.
  - Sagas for `OrderPlaced → Payment → Inventory`.
  - Outbox + events for reliability.
  - Idempotency, retries, and eventual consistency.

- **“How would you ensure consistency between services without distributed transactions?”**
  - Local transactions + Sagas.
  - Outbox pattern and idempotent consumers.
  - Clear state machine: `PENDING`, `CONFIRMED`, `CANCELLED`.

- **“Explain how you handle failures between microservices.”**
  - Timeouts, retries with backoff + jitter.
  - Circuit breakers, bulkheads.
  - Rate limiting and load shedding.
  - Observability (logs/metrics/traces) to debug incidents.

- **“How would you add observability to a microservices system?”**
  - Correlation/trace IDs.
  - Structured logging with context.
  - Metrics (RED), SLOs/alerts.
  - Distributed tracing via OpenTelemetry-like setup.

- **“How do you handle authentication and authorization in microservices?”**
  - Auth at gateway (JWT/opaque tokens).
  - Propagate identity via headers/claims.
  - Service-level authorization (RBAC/ABAC).
  - Service-to-service security (mTLS, mesh).

---

### 3. How to use this module

- Use this file as a **checklist**: can you explain each pattern in your own words and point to an implementation/demo in this repo?
- For any pattern you feel weak on, jump back to the corresponding module’s `.md` notes and Go code.
- Practice designing small systems verbally using:
  - Clear boundaries.
  - A couple of communication patterns.
  - 1–2 consistency patterns (Saga/Outbox).
  - Basic observability and security.

