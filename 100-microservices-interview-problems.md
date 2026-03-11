# 100 Microservices Interview Problems & Solutions

---

## 1. When should you *not* use microservices?

**Solution:** Prefer a monolith or modular monolith when the team is small, the domain is unclear, operational maturity is low, or traffic is modest. Microservices add network latency, more failure modes, and deployment/debugging complexity—only adopt when the benefits (independent scaling, team autonomy, technology diversity) outweigh these costs.

---

## 2. How do you decide service boundaries?

**Solution:** Use business capabilities and bounded contexts; optimize for independent deployability and data ownership. Avoid splitting purely by technical layers (e.g., “API service” vs “DB service”). Align boundaries with team ownership (Conway’s Law): if two teams can own and deploy parts independently, those are good candidates for separate services.

---

## 3. How do you prevent a “distributed monolith”?

**Solution:** Minimize synchronous call chains; enforce bounded contexts and avoid shared databases. Prefer async/event-driven integration where appropriate. Limit cross-service chattiness; keep APIs coarse-grained. Ensure each service can be deployed and scaled independently without requiring coordinated releases.

---

## 4. What does “service owns its data” mean?

**Solution:** Only that service reads from and writes to its datastore. Other services never access that database directly; they get data via APIs or events. Any replicated view (e.g., read models) is built from events or explicit APIs, not by joining across service DBs.

---

## 5. REST vs gRPC between services—what do you choose?

**Solution:** gRPC for internal service-to-service: strong typing, code generation, streaming, lower overhead. REST for public or polyglot consumers and when you want broad compatibility and simple tooling. Often: REST at the edge, gRPC internally.

---

## 6. How do you version APIs without breaking clients?

**Solution:** Prefer backward-compatible changes (add fields, don’t rename or remove). Use expand/contract: support both old and new during migration, then remove old. Version in URL path or header if needed. Consumer-driven contract tests catch breaking changes before release.

---

## 7. How do you handle partial failures?

**Solution:** Timeouts on every outbound call; retries with exponential backoff and jitter; circuit breakers to fail fast when a dependency is down; bulkheads to isolate resource usage; graceful degradation or fallbacks where acceptable; load shedding under overload. Observability (logs, metrics, traces) to detect and debug.

---

## 8. Why are timeouts mandatory on outbound calls?

**Solution:** Without timeouts, a slow or hung downstream can exhaust connections, goroutines, or threads and cause cascading failure. Timeouts bound the impact of a single bad dependency and keep the system responsive.

---

## 9. When are retries harmful?

**Solution:** When the failure is persistent (you amplify load on a failing service), when the operation is not idempotent (risk of double side-effects), or when there is no backoff/jitter (thundering herd). Also harmful when retry budgets are unbounded.

---

## 10. What is idempotency and why does it matter?

**Solution:** An operation is idempotent if performing it multiple times has the same effect as once. Matters because retries, duplicate messages, and at-least-once delivery are normal in distributed systems; idempotency lets you safely retry or replay.

---

## 11. Implement idempotency for `POST /payments`.

**Solution:** Require an `Idempotency-Key` header (or body field). Store the result of the first successful request keyed by that value (with TTL and uniqueness per client). On duplicate key, return the stored response without re-executing the charge. Ensure the actual charge is done once (e.g., via idempotency at the payment provider).

---

## 12. At-least-once messaging causes duplicates—how to handle?

**Solution:** Design consumers to be idempotent: track processed message IDs (inbox) or make the write idempotent (e.g., upsert with unique constraint). Use idempotency keys for operations that have side effects. Accept that duplicates may arrive; handle them in application logic.

---

## 13. Exactly-once semantics—do you promise it?

**Solution:** Usually no; true exactly-once is expensive and complex. Aim for “effectively once” via idempotent producers/consumers, deduplication, and transactional outbox where applicable. Be clear with stakeholders about “no duplicate processing” rather than “exactly one delivery.”

---

## 14. What is the circuit breaker pattern?

**Solution:** A component that tracks failures (or latency) for a dependency. After a threshold, it “opens” and fails fast without calling the dependency. After a cooldown, it goes “half-open” and allows a probe request; on success it “closes” again. Protects the system from cascading failure and gives the dependency time to recover.

---

## 15. Bulkhead pattern—give a concrete example.

**Solution:** Isolate resources per dependency. Example: separate connection pools or worker pools for “payments” vs “inventory” so that if payments is slow, it doesn’t exhaust all threads; inventory can still serve. In Go, limit concurrency per downstream with semaphores or dedicated clients.

---

## 16. Rate limiting strategies at the gateway.

**Solution:** Token bucket or leaky bucket per API key, user, or IP. Return 429 with `Retry-After` when exceeded. Apply different limits per route (e.g., stricter for write endpoints). Optionally use sliding window or distributed rate limiting (e.g., Redis) for consistency across gateway instances.

---

## 17. Load shedding vs rate limiting.

**Solution:** Rate limiting caps how much traffic you accept (e.g., 100 req/s per user). Load shedding is dropping or rejecting work when the system is already overloaded to protect core functions (e.g., reject non-critical requests when CPU or queue depth is high).

---

## 18. Design an API gateway for 10 services.

**Solution:** Single entry point; routing by path or host to backend services; authentication/authorization at the edge; rate limiting and request/response normalization; centralized logging and tracing. Keep business logic in services; gateway is infrastructure. Use config-driven routing and health checks for backends.

---

## 19. BFF vs API gateway—difference?

**Solution:** API gateway is generic edge infrastructure (routing, auth, rate limit). BFF (Backend for Frontend) is a client-specific backend that aggregates and shapes data from multiple services for one client type (e.g., mobile app). A BFF can sit behind a gateway or be the gateway for that client.

---

## 20. Aggregation endpoint causes N+1 calls—fix it.

**Solution:** Add batch endpoints in downstream services, cache where appropriate, or fan out in parallel with a single timeout budget. Alternatively, maintain a read model (CQRS) that is pre-aggregated so the BFF does one read. Avoid sequential calls in a loop; use bulk APIs or async aggregation.

---

## 21. Service discovery options.

**Solution:** DNS-based (e.g., Kubernetes Services), client-side discovery with a registry (Consul, Eureka), or server-side (load balancer with health checks). In Kubernetes, DNS is the default; in non-K8s environments, a registry plus health checks is common.

---

## 22. How does Kubernetes service discovery work?

**Solution:** You define a Service (name, selector). Cluster DNS resolves that name to a virtual IP. kube-proxy (or the CNI) forwards traffic to healthy pods matching the selector. No application code needed; use the service name as hostname.

---

## 23. Config management across environments.

**Solution:** 12-factor: config via environment variables or config files, not baked into the image. Use ConfigMaps/Secrets in K8s; or a config server with environment-specific overrides. Never commit secrets; use a secret manager. Support overrides (env > file) and sensible defaults.

---

## 24. Secret management best practices.

**Solution:** Never commit secrets; use a secret store (e.g., Kubernetes Secrets with encryption at rest, HashiCorp Vault, cloud KMS). Rotate secrets periodically; use short-lived tokens where possible. Inject at runtime; restrict access with RBAC. Avoid logging or exposing secrets in errors.

---

## 25. Health check vs readiness check.

**Solution:** Liveness: “Is the process alive?”—if not, restart. Readiness: “Should this instance receive traffic?”—if not, remove from load balancer. Readiness can fail during startup or when a critical dependency is down so traffic is not sent until the service is ready.

---

## 26. Graceful shutdown in a Go microservice.

**Solution:** Listen for SIGTERM; stop accepting new connections (`Server.Shutdown(ctx)`); allow in-flight requests to complete within a timeout; then exit. Use `signal.NotifyContext` or similar to trigger shutdown and pass a deadline context to `Shutdown`.

---

## 27. Observability pillars—what do you instrument first?

**Solution:** Metrics first: request rate, error rate, latency (RED). Then structured logs with correlation IDs. Then distributed tracing for cross-service flows. Alerts should be SLO-based (e.g., error budget burn rate) rather than raw thresholds where possible.

---

## 28. What is a trace and a span?

**Solution:** A trace is the end-to-end journey of a request. A span is a single unit of work (e.g., an HTTP call, DB query) with start/end time and attributes. Spans are linked (parent/child) to form a trace tree. Use trace ID to correlate logs and metrics.

---

## 29. Implement correlation IDs.

**Solution:** Middleware reads `X-Request-ID` (or similar) from the request or generates one. Store in context; log it on every log line; set the same header on outbound HTTP/gRPC calls. Downstream services continue the chain so the whole path is correlated.

---

## 30. SLO / SLI / error budget—explain.

**Solution:** SLI = measurable indicator (e.g., availability, latency). SLO = target (e.g., 99.9% availability). Error budget = allowed failure (e.g., 0.1% of requests). Use error budget to decide when to pause releases or invest in reliability. Alerts when burn rate consumes budget too fast.

---

## 31. Debugging high p99 latency.

**Solution:** Use distributed tracing to find which service or span is slow. Check for saturation (CPU, DB, GC), lock contention, slow queries, or retry storms. Look for “tail amplification” (e.g., timeouts causing retries that multiply load). Fix the bottleneck or add caching/parallelism.

---

## 32. Retry storm incident—what do you do?

**Solution:** Immediately: disable or reduce retries, open circuit breakers, or rate limit at the edge to stop hammering the failing dependency. Then fix the root cause (bug, capacity, dependency). Use backoff and jitter to prevent synchronized retries. Consider a global retry budget.

---

## 33. Distributed transactions across services—why avoid?

**Solution:** Two-phase commit (2PC) across services creates tight coupling, single points of failure, and blocking. Use sagas instead: sequence of local transactions with compensating actions. Accept eventual consistency and design for it (idempotency, clear state machine).

---

## 34. Saga orchestration vs choreography.

**Solution:** Orchestration: a central coordinator tells each service what to do and handles compensations. Easier to reason about and modify. Choreography: each service reacts to events and triggers the next step; no central brain but flow is distributed and can be harder to follow. Use orchestration when the flow is complex or central control is desired.

---

## 35. Design order placement with payment and inventory.

**Solution:** Saga: create order (PENDING), reserve inventory, authorize payment, then confirm. On any failure, run compensations: release inventory, cancel order, refund if needed. Use idempotent steps and idempotency keys for payment. Persist saga state for recovery.

---

## 36. Outbox pattern—what does it fix?

**Solution:** Ensures “write to DB” and “publish event” are atomic without distributed transactions. Write the event to an outbox table in the same transaction as the business write. A separate process polls the outbox and publishes; mark as published. Guarantees at-least-once publishing and no “write succeeded but publish failed” inconsistency.

---

## 37. Inbox pattern—why needed?

**Solution:** For idempotent consumption of messages. Store incoming message IDs (and optionally payload) in an inbox table before processing. Skip or return cached result if already seen. Handles duplicate deliveries and allows safe replay.

---

## 38. CQRS—when is it justified?

**Solution:** When read and write workloads differ significantly (e.g., many different read shapes, high write load, or need for optimized read models). Also when you want to scale reads independently. Avoid if it adds unnecessary complexity for simple CRUD.

---

## 39. Materialized view design for “user order history”.

**Solution:** Consume order lifecycle events; maintain a read model keyed by user_id (e.g., in a table or search index). Update on OrderCreated, OrderCancelled, etc. Queries hit this view. Accept eventual consistency; use version or timestamp for ordering.

---

## 40. Event sourcing—benefits and tradeoffs.

**Solution:** Benefits: full audit trail, replay, temporal queries, flexible read models. Tradeoffs: event schema evolution, storage growth, complexity in querying and debugging. Use when audit and replay are first-class requirements.

---

## 41. Handling event schema evolution.

**Solution:** Version events (e.g., `OrderCreated_v2`); keep backward compatibility (add optional fields, don’t remove). Consumers handle multiple versions or use an upcaster. Never reuse field semantics; deprecate and add new. Contract test event schemas.

---

## 42. Contract testing approach.

**Solution:** Consumers define expectations (e.g., Pact): “When I send this request, I expect this response.” Provider runs these tests in CI. Catches breaking changes before deployment. Run consumer tests against provider stub and provider tests against real or mock consumer contracts.

---

## 43. Blue/green vs canary deployments.

**Solution:** Blue/green: two identical environments; switch traffic all at once; instant rollback by switching back. Canary: route a small percentage of traffic to the new version; increase gradually; rollback by shifting traffic back. Canary reduces blast radius but requires good metrics and automation.

---

## 44. Feature flags—how to use safely?

**Solution:** Default new features off; use flags for gradual rollout and kill switch. Avoid long-lived flags; clean up after rollout. Control rollout by percentage, user segment, or environment. Audit who can change flags; avoid flags for security-critical paths unless carefully gated.

---

## 45. Rolling back a breaking release.

**Solution:** Revert the deployment (e.g., rollback to previous K8s revision). Ensure database migrations are backward compatible (expand/contract); if you already ran a migration, the old code must still work with the new schema or you need a separate rollback migration. Use feature flags to disable new behavior if possible.

---

## 46. Expand/contract database migration example.

**Solution:** Expand: add new column as nullable or with default; deploy code that writes both old and new. Backfill. Contract: deploy code that reads only from new column; then remove old column in a later release. Never remove in the same release as the one that stops writing to it.

---

## 47. Authentication patterns for microservices.

**Solution:** Authenticate at the edge (API gateway or BFF) via JWT or opaque token. Propagate identity to downstream services (e.g., token relay or signed claims in headers). Each service can validate tokens or trust the gateway. For service-to-service, use mTLS or service tokens.

---

## 48. JWT vs opaque token—choose one.

**Solution:** JWT: stateless validation, good for distributed systems; revocation is harder. Opaque: validated via introspection; easy to revoke. Use JWT for internal services with short TTL; use opaque when revocation and central control matter (e.g., session tokens).

---

## 49. Token revocation with JWT.

**Solution:** Short expiry plus refresh tokens; refresh tokens stored and revocable. For immediate revocation: maintain a denylist (e.g., Redis) of revoked jti or token IDs; check on each request. Or rotate signing keys and invalidate all tokens signed with the old key.

---

## 50. mTLS in a service mesh—what does the app do?

**Solution:** Typically nothing. The sidecar proxy handles TLS; the app speaks plain HTTP to localhost. The mesh injects certificates and enforces policy. App may need to trust the proxy and forward client identity (e.g., from header) if required for authorization.

---

## 51. Preventing PII leakage in logs.

**Solution:** Don’t log full request/response bodies by default; allowlist fields that are safe. Redact or omit sensitive fields (passwords, tokens, PII). Use structured logging with explicit fields. Consider log levels and access control; comply with data classification policies.

---

## 52. Multi-tenant microservice—how to isolate tenants?

**Solution:** Tenant ID in auth context; enforce at data access (every query filtered by tenant). Optionally separate DB or schema per tenant for stronger isolation. Rate limit and quota per tenant. Audit tenant-scoped actions. Never allow cross-tenant data access.

---

## 53. API rate limit per user and per route.

**Solution:** Use a composite key (e.g., `user_id` + `route` or `method+path`) in a token bucket or sliding window. Enforce at gateway; return 429 and optionally `Retry-After`. Configure different limits per route (e.g., stricter for write, looser for read).

---

## 54. Preventing cascading failures across a 5-service call chain.

**Solution:** Introduce timeouts at each hop (and an overall deadline). Use circuit breakers and bulkheads so one failing service doesn’t exhaust callers. Consider async boundaries (queue) to decouple. Cache where appropriate. Implement fallbacks or default responses for non-critical path.

---

## 55. Thundering herd on cache miss.

**Solution:** Request coalescing (singleflight): only one request per key fills the cache; others wait. Stale-while-revalidate: serve stale and refresh in background. Add jitter to TTLs. Pre-warm cache for hot keys. Use a lock or semaphore per key if necessary.

---

## 56. Cache invalidation for product price updates.

**Solution:** On price change, publish an event or call an invalidation API. Invalidate by key or pattern (e.g., product ID). Optionally use versioned keys (e.g., `price:v123`) so old requests can still complete. TTL as a safety net. Consider write-through for critical paths.

---

## 57. Strong consistency requirement across services—what do you do?

**Solution:** First question: can the invariant live in one service? If yes, move it. If not, consider saga with clear state and compensations; or accept eventual consistency with well-defined semantics (e.g., read-your-writes via same service). Avoid 2PC unless absolutely necessary; document the tradeoff.

---

## 58. Hot partition/key in database.

**Solution:** Redistribute writes (e.g., add a random suffix to the key, or use a different sharding key). Cache the hot key aggressively. Queue writes and batch. If one tenant is hot, consider separate handling or rate limiting. Design partition key to avoid skew.

---

## 59. Designing for multi-region.

**Solution:** Decide data residency and replication (sync vs async). Use routing (e.g., latency-based) to direct users to nearest region. Accept eventual consistency for cross-region data; use conflict resolution (e.g., last-write-wins, version vectors). Design for idempotency and replay in case of duplication.

---

## 60. Handling clock skew in distributed systems.

**Solution:** Don’t rely on wall-clock time for ordering or uniqueness. Use logical clocks, version vectors, or hybrid logical clocks. For timestamps, use server-provided time or accept approximate ordering. For IDs, use time-based but include enough entropy (e.g., ULID).

---

## 61. Generate unique IDs across services.

**Solution:** Use UUID v4 for randomness or ULID/UUIDv7 for time-ordered uniqueness without a central coordinator. Avoid DB auto-increment across services. If you need sortability and uniqueness, use a combination of timestamp and node ID plus sequence, or a dedicated ID service with high availability.

---

## 62. Observability: what to alert on?

**Solution:** SLO-based alerts (e.g., error budget burn rate). Error rate spike, latency degradation (e.g., p95), dependency failure rate. Resource saturation (CPU, memory, queue depth). Avoid alerting on every blip; use windows and severity. Prefer few actionable alerts.

---

## 63. Designing a dead-letter queue strategy.

**Solution:** After N retries, move message to DLQ with metadata (reason, count, original message). Provide tooling to inspect, republish, or discard. Ensure consumers are idempotent so republish is safe. Consider retention and access control for DLQ.

---

## 64. Poison message handling.

**Solution:** Detect repeated failures for the same message (e.g., by message ID). After threshold, move to DLQ and stop retrying. Optionally quarantine and alert. Fix the bug or schema; then replay from DLQ with a fix. Validate early to avoid poisoning the pipeline.

---

## 65. Ordering guarantees in event processing.

**Solution:** Partition by a key (e.g., order_id) so all events for that key go to one partition; process sequentially per partition. Include sequence or version in the event. If strict order is not required, accept eventual consistency and handle out-of-order (e.g., by version or timestamp).

---

## 66. Exactly-once payment capture in an async system.

**Solution:** Use idempotency key at the payment provider; store it with the order/payment state. Deduplicate incoming events (inbox). Ensure the “capture” step is idempotent and that you never double-capture. Use outbox for emitting “payment captured” only after local commit.

---

## 67. Backpressure in async pipelines.

**Solution:** Use bounded queues; when full, block or reject producers (or slow them down). Scale consumers to match throughput. Expose queue depth and consumer lag as metrics; alert and auto-scale. Implement rate limiting or flow control on producers if needed.

---

## 68. Service mesh retries vs application retries—who owns it?

**Solution:** Prefer a single place to avoid double retries (e.g., mesh retries + app retries = too many). If mesh does retries, app must be idempotent and use reasonable retry policy. Document and configure consistently; often mesh handles transport retries, app handles business retries or vice versa. Set budgets and disable one layer if redundant.

---

## 69. Handling large payloads between services.

**Solution:** Prefer passing a reference (e.g., URL to object storage) instead of the full payload. Use streaming (e.g., gRPC streaming, chunked upload). Enforce max body size at gateway; return 413 for oversized. Compress when appropriate. Avoid putting large payloads in message queues; use references.

---

## 70. API pagination strategy.

**Solution:** Cursor-based pagination for large or changing datasets (stable, efficient). Return an opaque cursor for “next page.” Offset-based only for small or static data. Include total count only if cheap; otherwise avoid. Use consistent sort order.

---

## 71. Designing a “search” service.

**Solution:** CQRS: write side emits events; search service consumes and updates an index (e.g., Elasticsearch). Queries hit the index only. Accept eventual consistency. Support reindexing (full or by time window) for recovery. Consider ranking, filters, and access control.

---

## 72. Dealing with synchronous fanout to many services.

**Solution:** Call in parallel with a single deadline; aggregate results. Tolerate partial failure (return what you have or fail fast). Consider caching or precomputed views to avoid fanout on every request. Limit fanout (e.g., max N services) or move to async aggregation.

---

## 73. How to do distributed locking?

**Solution:** Use a store that supports it (e.g., Redis SET NX with TTL, etcd lease). Always set a timeout so locks are released. Handle lock loss (e.g., lease expiry) and ensure work is idempotent or use a fencing token. Prefer avoiding distributed locks; use optimistic concurrency or single-writer design where possible.

---

## 74. Optimistic concurrency control example.

**Solution:** Store a version (or timestamp) with the entity. On update, send the version; the server checks it matches and increments. If mismatch, return 409; client refetches and retries. Prevents lost updates under concurrent writes without locking.

---

## 75. Multi-step workflow with human approval.

**Solution:** Model as a state machine; persist state in DB. Use saga for the automated steps; “human approval” is a step that completes asynchronously (e.g., via callback or polling). Use timeouts and compensations for stuck approvals. Emit events for auditing and downstream consumers.

---

## 76. Preventing “shared library coupling” across services.

**Solution:** Share only minimal, stable contracts (e.g., API specs, event schemas) and perhaps small utility libraries. Avoid sharing domain logic or fat client libraries that force upgrades. Prefer code generation from OpenAPI/protobuf so each service owns its client. Version contracts and support backward compatibility.

---

## 77. Handling breaking change in an event consumer.

**Solution:** Version the event; consumer supports both old and new (or multiple versions) during transition. Deploy consumer first if it can handle old format; then deploy producer with new format. Use schema registry and compatibility checks. Run dual-write or expand/contract if needed.

---

## 78. Multi-API clients and backward compatibility.

**Solution:** Add new endpoints or fields; keep old ones until deprecation period. Use versioning (URL or header). Contract tests for consumers. Document deprecation and sunset. Provide migration path and avoid breaking changes without notice.

---

## 79. Canary analysis—what metrics do you compare?

**Solution:** Error rate, latency (e.g., p50, p95, p99), saturation (CPU, memory), and optionally business metrics (conversion, revenue). Use statistical significance (e.g., enough sample size) and define rollback criteria in advance (e.g., error rate +2%). Automate canary evaluation and rollback.

---

## 80. Observability for async processing.

**Solution:** Propagate trace context in message headers; create spans for each processing step. Emit metrics: queue depth, processing lag, processing time, error rate per consumer. Log with correlation ID. Use dead-letter and replay tooling with observability in mind.

---

## 81. Handling “stuck” sagas.

**Solution:** Persist saga state; use timeouts to trigger compensation or retry. Implement retry with backoff for transient failures. Provide operator tooling to list stuck sagas, view state, and manually trigger compensation or resume. Ensure steps are idempotent so retries are safe.

---

## 82. Compensating transaction design.

**Solution:** For each forward step, define a compensating action (e.g., “reserve inventory” → “release reservation”). Compensations must be idempotent and safe to run multiple times. Order of compensation is reverse of forward steps. Log and monitor compensations for debugging.

---

## 83. How do you do “read your writes” across services?

**Solution:** Route the read to the same service that did the write when possible. Or use session/causal consistency (e.g., pass a token that the read side waits for). Alternatively, read from the write-side store until the read model is updated (e.g., CQRS with same-service read).

---

## 84. Preventing inconsistent UI after eventual consistency.

**Solution:** Show optimistic UI (assume success) and reconcile; or show “pending” until confirmed. Poll or use WebSockets for updates. Read from the same service that accepted the write when possible. Set user expectations (e.g., “changes may take a moment to appear”).

---

## 85. API gateway becomes bottleneck—what next?

**Solution:** Scale gateway horizontally (stateless). Offload heavy aggregation to BFFs or backend services. Cache responses where appropriate. Split gateways by domain or client. Use CDN for static assets. Profile and optimize hot paths; consider connection pooling and timeouts.

---

## 86. Testing microservices locally.

**Solution:** Use Docker Compose to run dependencies (DB, queue, optional services). Run the service under test in IDE or container. Use contract tests and stub external services. Seed test data; use test config and feature flags. Optionally use ephemeral environments or service virtualization.

---

## 87. Chaos testing plan for checkout flow.

**Solution:** Inject latency and failures into payment and inventory services; kill pods; drop messages. Verify: orders are not confirmed without payment; inventory is not oversold; compensations run on failure; no double charge. Start with small blast radius; automate and run in staging first.

---

## 88. Secure internal APIs from lateral movement.

**Solution:** mTLS and service identity (e.g., SPIFFE); authorize every call (zero trust). Use network policies to restrict which pods can talk. Least-privilege IAM and RBAC. Audit service-to-service calls. Don’t rely on “internal network” as trusted.

---

## 89. Handling SSRF risk in microservices.

**Solution:** Validate and allowlist URLs that the service can call; block metadata and internal IP ranges. Use a proxy or egress firewall. Don’t pass user-controlled URLs directly to outbound HTTP client. Apply principle of least privilege for network access.

---

## 90. Preventing replay attacks with tokens.

**Solution:** Use TLS to prevent token theft in transit. Short-lived tokens; bind to client (e.g., fingerprint) for sensitive operations. Include `iat` (issued at) and validate. For critical actions, use nonce or one-time tokens. Rotate signing keys and revoke compromised tokens.

---

## 91. Per-service database vs shared database tradeoff.

**Solution:** Per-service DB: autonomy, independent scaling, clear ownership; but no joins across services and eventual consistency. Shared DB: simpler queries and strong consistency but couples services and deployments. Prefer per-service for true microservices; use events/APIs for cross-service data.

---

## 92. How do you implement service-to-service authorization?

**Solution:** Use service identity (e.g., from mTLS or JWT with service account). Policy engine (e.g., OPA) to decide if service A can call endpoint X on service B. Enforce in sidecar or gateway. Map identity to roles or attributes; audit decisions.

---

## 93. Observability: why structured logs instead of plain text?

**Solution:** Structured (e.g., JSON with fields) is queryable and aggregatable; easy to filter by trace_id, level, service. Enables alerting and integration with log analytics. Plain text is harder to parse and correlate at scale.

---

## 94. Handling high-cardinality metrics.

**Solution:** Avoid high-cardinality dimensions (e.g., user_id) in metrics; use sampling or aggregation. Put high-cardinality data in logs or traces. Use metrics for rates and histograms with bounded dimensions (e.g., route, status code). Otherwise cost and performance of metrics store can explode.

---

## 95. Choose between queue vs stream (Kafka-like).

**Solution:** Queue: task distribution; message deleted after ack; good for work queues. Stream: ordered, replayable log; multiple consumers; good for event sourcing and replay. Choose stream when you need ordering, replay, or multiple consumer groups; queue for simple work distribution.

---

## 96. How to replay events safely.

**Solution:** Consumers must be idempotent (e.g., by id or version). Use a separate consumer group or flag for replay so you don’t affect production. Replay to a new table or with a version bump to avoid overwriting current state incorrectly. Monitor replay progress and lag.

---

## 97. Handling “exactly once” business requirement from stakeholders.

**Solution:** Clarify the invariant: usually “no double charge” or “no duplicate order,” not “exactly one message delivery.” Implement effectively-once processing with idempotency keys, idempotent consumers, and deduplication. Document that we guarantee no duplicate *effect*, not exactly-once delivery.

---

## 98. Debug “works in staging, fails in prod” microservices.

**Solution:** Compare config, secrets, and feature flags between envs. Check for traffic patterns, timeouts, and resource limits. Verify dependency versions and data shape. Use distributed tracing and logs to compare flows. Reproduce in staging with prod-like data and load; add temporary logging if needed.

---

## 99. Dependency timeout tuning across services.

**Solution:** Set an overall deadline at the entry and propagate. Each downstream timeout should be less than the caller’s remaining budget. Allocate timeout budget per hop (e.g., 80% to dependency, 20% buffer). Avoid infinite retries; use retry budget. Monitor tail latency and adjust.

---

## 100. “Design a microservices platform” for interviews.

**Solution:** Describe a paved path: standard service template (layout, config, health, tracing). CI/CD (build, test, contract tests, deploy). Observability (logs, metrics, traces, SLOs). Service discovery and config. Security (authn/z, mTLS, secrets). Runtime (e.g., Kubernetes). API gateway and optionally mesh. Messaging and events. Incident response and runbooks. Emphasize consistency and developer experience over each team building everything from scratch.

---

*End of 100 Microservices Interview Problems & Solutions*
