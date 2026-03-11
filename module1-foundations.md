## Module 1 – Foundations (Concepts + Minimal Go Service)

### 1. What microservices are (and are not)

- **Microservice**: A small, independently deployable service that owns its data and implements a specific business capability.
- Compared to:
  - **Monolith**: Single deployable unit (one codebase, one process). Simple to start, harder to evolve at scale.
  - **Modular monolith**: Single deployable, but internally well-modularized. Often a good stepping stone before microservices.
- **Key properties of a microservice**
  - Independent deployability
  - Clear, narrow responsibility (single business capability)
  - Owns its own data store
  - Communicates with others through well-defined APIs

**When microservices are a bad idea**

- Small team, simple product, limited scale → monolith or modular monolith is usually better.
- If you don’t have:
  - Good observability
  - Automation (CI/CD, infra)
  - Operational experience
  then microservices add a lot of complexity.

Interview angle: Be ready to say **“microservices are an organizational and operational choice, not just an architecture style”** and explain Conway’s Law.

---

### 2. Distributed-systems fundamentals you must internalize

- **Latency**: Every network hop adds delay. Treat remote calls as *slow* and *unreliable*.
- **Partial failure**: In a distributed system, some components can fail while others continue running.
- **Retries & timeouts**:
  - Always set timeouts on outbound calls.
  - Retries must have backoff + jitter to avoid retry storms.
- **Idempotency**:
  - Doing the same operation multiple times should ideally produce the same effect (e.g., charging a card once even if the request is retried).
- **Ordering and duplication**:
  - Don’t rely on exact ordering across services.
  - Design for at-least-once delivery—handle duplicates.
- **Consistency (very high-level for now)**:
  - Strong consistency: reads see the latest writes.
  - Eventual consistency: reads may see stale data briefly, but will converge.

Mental rule: **“Calls inside a process are almost free; calls over the network are expensive and can fail.”**

---

### 3. Communication basics: HTTP, REST, gRPC (at a glance)

- **HTTP/REST**
  - Resources (nouns): `/users`, `/orders/{id}`
  - Verbs: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
  - Status codes: `2xx` success, `4xx` client error, `5xx` server error.
- **gRPC** (conceptually)
  - Uses HTTP/2 and Protobuf.
  - Strong typing, code generation.
  - Good for service-to-service communication, streaming.
- **Sync vs async**
  - Synchronous: request–response; client waits.
  - Asynchronous: messages/events; client doesn’t block waiting.

In Module 1, we focus on **HTTP** with JSON, because that’s the default for many interviews and easy to run.

---

### 4. Go for microservices – key primitives

- **`http.Server` + handlers**: For building HTTP APIs.
- **`context.Context`**:
  - Carries deadlines, cancelation signals, and metadata (like correlation IDs).
  - Flows from incoming request to any work you do.
- **Graceful shutdown**:
  - Stop accepting new requests.
  - Finish in-flight requests within a timeout.
- **Configuration via env** (12-factor style):
  - Use environment variables for ports, DB URLs, feature flags, etc.

We will now implement a **minimal Go HTTP service** that demonstrates:

- Health and readiness endpoints.
- Config-driven port.
- Structured logging (simple).
- Context usage.
- Graceful shutdown with OS signals.

---

### 5. Interview checklist from Module 1

You should be able to answer:

- Why might you **not** choose microservices for a new startup project?
- What does it mean that a service “owns its data”?
- Why are timeouts and retries important in distributed systems?
- What is idempotency, and why is it important for retries?
- How would you implement graceful shutdown in a Go HTTP service?

If any of these are unclear after reading and coding, ask for clarification before moving on.

