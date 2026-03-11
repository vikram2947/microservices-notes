## Microservices Mastery Roadmap (Go, Pattern-First – Path B)

This roadmap defines everything we will cover, in order, with a **pattern-first**, **Go-based** approach. Each module will have:

- **Notes** (concepts, patterns, interview angles)
- **Runnable Go code** (focused mini-projects)
- **Exercises / interview questions**

After finishing each module, I will ask you whether to **move ahead**.

---

## Module 1 – Foundations (Concepts + Minimal Service)

- **Microservices basics**
  - What microservices are (and are not)
  - Monolith vs modular monolith vs microservices
  - Conway’s Law and team boundaries
  - When microservices hurt more than help
- **Distributed-systems fundamentals**
  - Latency, partial failure, timeouts, retries, backpressure
  - Idempotency, duplication, ordering basics
  - Basic consistency models (strong vs eventual)
- **Protocols and communication basics**
  - HTTP/REST fundamentals, status codes, resource modeling
  - gRPC and Protobuf basics (conceptual intro)
  - Synchronous vs asynchronous communication
- **Go for services**
  - Project layout and configuration (12-factor-ish)
  - HTTP server, routing, handlers, middleware
  - Context propagation, graceful shutdown

**Deliverables**
- Minimal Go HTTP microservice:
  - Health checks, readiness, liveness
  - Basic logging and configuration via env
  - Graceful shutdown with `context.Context`

---

## Module 2 – Service Design & Boundaries (DDD-Lite)

- **Decomposition**
  - When to split a service
  - Business capabilities and subdomains
  - Data ownership and avoiding shared databases
- **Domain-Driven Design (only what we need)**
  - Bounded Contexts
  - Aggregates and invariants
  - Domain Events (as a modeling tool)
- **API and contract design**
  - Resource vs operation-oriented APIs
  - Versioning strategies and backward compatibility
  - Pagination, filtering, idempotency keys

**Deliverables**
- 2–3 small Go services (e.g., `users`, `orders`, `payments`) with:
  - Clear API contracts (OpenAPI / protobuf definitions)
  - Separate data models and ownership

---

## Module 3 – Communication Patterns (Sync/Async) & Reliability

- **Synchronous communication patterns**
  - Request–response
  - API composition/aggregation
  - Fan-out/fan-in patterns
  - gRPC unary and deadlines (concept + small demo)
- **Asynchronous communication patterns**
  - Event-driven architecture basics
  - Pub/sub vs queues vs streams
  - Competing consumers, message ordering strategies
- **Reliability and resilience**
  - Timeouts, retries with backoff and jitter
  - Circuit breaker pattern
  - Bulkhead pattern
  - Rate limiting and load shedding basics
- **Idempotency and deduplication**
  - Idempotency keys in HTTP
  - At-least-once delivery and idempotent consumers

**Deliverables**
- Go examples:
  - HTTP call between services with timeouts and retries
  - A simple message-driven flow using a local queue (e.g., in-memory or Redis-like)
  - Circuit-breaker and rate-limiter middleware

---

## Module 4 – Data, Transactions & Consistency Patterns

- **Distributed transactions and why 2PC is avoided**
- **Consistency models**
  - Strong vs eventual consistency (practical implications)
  - Read-your-writes, monotonic reads (conceptual)
- **Key data patterns**
  - Saga pattern
    - Orchestration vs choreography
  - Outbox pattern
  - Transactional messaging and deduplication
  - CQRS (Command Query Responsibility Segregation)
  - Materialized views and read models
  - Event sourcing (concept + minimal example)
- **Concurrency control**
  - Optimistic locking (versioning)
  - Conflict resolution strategies

**Deliverables**
- Go implementation:
  - A small Saga across 2–3 services (orchestration-based)
  - Outbox pattern in one service (with background publisher)
  - Simple CQRS + read model projection

---

## Module 5 – API Gateway, BFF & Edge Patterns

- **API Gateway**
  - Responsibilities: routing, auth, rate limiting, aggregation
  - What should and should not be in a gateway
- **Backend-for-Frontend (BFF)**
  - Why BFFs exist
  - Tailoring APIs per client (web, mobile)
- **Gateway-related patterns**
  - Edge authentication and authorization
  - Request/response transformation
  - Cross-cutting concerns (logging, tracing, metrics)

**Deliverables**
- Go-based edge:
  - Simple API gateway/BFF that routes to multiple backend services
  - Basic auth and rate limiting at the edge

---

## Module 6 – Service Discovery, Configuration & Networking

- **Service discovery**
  - Static vs dynamic discovery
  - DNS-based discovery (Kubernetes, etc.)
  - Conceptual overview of Consul/Eureka-style registries
- **Configuration & secrets**
  - 12-factor config
  - Externalized configuration (env, files, config service concept)
  - Secret management best practices (without deep vendor specifics)
- **Networking realities**
  - HTTP connection pooling, keep-alives
  - Latency sources, timeouts
  - TLS vs mTLS fundamentals

**Deliverables**
- Go services:
  - Config loading (env + config file)
  - Simple “service directory” and discovery simulation
  - TLS-enabled communication example

---

## Module 7 – Observability: Logging, Metrics, Tracing

- **Observability pillars**
  - Logs (structured, contextual)
  - Metrics (counters, gauges, histograms)
  - Traces (distributed tracing)
- **Patterns**
  - Correlation IDs and trace IDs
  - RED and USE metrics
  - SLOs, SLIs, error budgets (conceptual)
- **Debugging real issues**
  - Timeouts and retry storms
  - Hot endpoints and tail latency

**Deliverables**
- Go implementation:
  - Structured logging with correlation IDs
  - Metrics (Prometheus-style) for a few endpoints
  - Distributed tracing with OpenTelemetry across multiple services

---

## Module 8 – Security for Microservices

- **Authentication & authorization**
  - JWT, opaque tokens, basic flows
  - Role-based access control (RBAC) basics
- **Service-to-service security**
  - mTLS for service communication (concept + small demo)
  - Token propagation (token relay pattern)
- **API security considerations**
  - Rate limiting and throttling
  - Input validation and schema validation
  - Basic OWASP-style concerns for APIs

**Deliverables**
- Go services:
  - Auth service that issues tokens
  - Gateway enforcing auth and forwarding identity
  - Simple mTLS example between two services

---

## Module 9 – Containers, Orchestration & Kubernetes Basics

- **Containers**
  - Docker images, multi-stage builds, minimal images
  - Containerizing Go microservices
- **Kubernetes essentials**
  - Pods, Deployments, Services, Ingress
  - ConfigMaps, Secrets
  - Liveness/readiness/startup probes
  - Horizontal Pod Autoscaler basics
- **Deployment strategies**
  - Rolling updates
  - Blue/green and canary (concept + simple manifests)

**Deliverables**
- Containerized Go services:
  - Dockerfiles for services
  - Kubernetes manifests (YAML) to run them locally (e.g., with kind/minikube)

---

## Module 10 – CI/CD & Delivery Safety

- **Pipelines**
  - Build, test, lint, security scan, deploy
- **Testing strategies**
  - Unit, integration, component, and end-to-end tests
  - Consumer-driven contract testing
- **Database and schema evolution**
  - Expand/contract migrations
  - Backward/forward compatible schema design

**Deliverables**
- Example:
  - Simple CI pipeline definition (e.g., GitHub Actions-style YAML as a reference)
  - Contract tests between two Go services (consumer/provider)
  - Example of safe schema migration flow

---

## Module 11 – Performance, Scalability & Caching

- **Performance concepts**
  - Throughput, latency, tail latency, percentiles
  - Load patterns and capacity planning (high-level)
- **Caching strategies**
  - In-memory cache (per-instance)
  - Shared cache (Redis-like)
  - Cache invalidation patterns
- **Queues and buffering**
  - Using queues to smooth traffic spikes
  - Dead-letter queues (DLQ) for poison messages
- **Go performance tooling**
  - Profiling with `pprof`
  - Connection pooling and reuse

**Deliverables**
- Go examples:
  - Caching layer for a read-heavy endpoint
  - Simple load test + profiling
  - Queue-based “burst smoothing” example

---

## Module 12 – Service Mesh & Advanced Traffic Management (Optional but Powerful)

- **Service mesh concepts**
  - Sidecar proxies, data plane vs control plane
  - mTLS, retries, circuit breaking, timeouts in mesh
  - Traffic splitting and canary releases at mesh level
- **Tradeoffs**
  - Complexity vs benefits
  - When not to use a mesh

**Deliverables**
- Conceptual + (if your environment allows):
  - Minimal example mesh configuration for a couple of services (e.g., Istio/Linkerd-like concepts)

---

## Module 13 – Pattern Catalog & Interview Drilldown

We will ensure you understand and can explain (with examples/tradeoffs) at least the following patterns:

- **Decomposition & design**
  - Bounded Context
  - Strangler Fig
  - Anti-Corruption Layer
- **Communication**
  - API Gateway
  - Backend for Frontend (BFF)
  - Aggregator
  - Pub/Sub
  - Competing Consumers
- **Data & consistency**
  - Database-per-service
  - Saga (orchestration & choreography)
  - Outbox pattern
  - CQRS
  - Event Sourcing (concept + minimal demo)
  - Materialized View
  - Idempotent Consumer
  - Transactional Messaging
- **Resilience**
  - Circuit Breaker
  - Bulkhead
  - Timeout
  - Retry with Exponential Backoff & Jitter
  - Rate Limiter
  - Load Shedding
  - Dead Letter Queue
  - Fallback
- **Operations & delivery**
  - Service Discovery
  - Config Externalization
  - Health Checks
  - Blue/Green Deployments
  - Canary Releases
  - Feature Flags
- **Observability**
  - Correlation IDs
  - Structured Logging
  - Distributed Tracing
  - Metrics and SLIs
- **Security**
  - mTLS
  - Token Relay
  - Zero-Trust-inspired service communication (conceptual)
- **Testing**
  - Contract Testing
  - Chaos Testing (concepts, not necessarily tooling)

**Deliverables**
- Consolidated notes and short Go snippets for each pattern where applicable.
- Interview-style questions and “gold standard” answer outlines.

---

## Learning Approach

- We follow **Path B: pattern-first**:
  - Each module focuses on a set of patterns and concepts.
  - Each module has focused, runnable Go examples.
  - We connect patterns across modules as we progress.
- After **each module**:
  - You confirm when you are ready to move ahead.
  - I can quiz you with interview-style questions if you want.

